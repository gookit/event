package event

import (
	"reflect"
	"strings"
	"sync"
)

// Manager event manager definition. for manage events and listeners
type Manager struct {
	Options
	sync.Mutex

	ch   chan Event
	once sync.Once

	// name of the manager
	name string
	// pool sync.Pool
	// is a sample for new BasicEvent
	sample *BasicEvent

	// storage user custom Event instance. you can pre-define some Event instances.
	events map[string]Event
	// storage user pre-defined event factory func.
	eventFc map[string]FactoryFunc

	// storage all event name and ListenerQueue map
	listeners map[string]*ListenerQueue
	// storage all event names by listened
	listenedNames map[string]int
}

// NewManager create event manager
func NewManager(name string, fns ...OptionFn) *Manager {
	em := &Manager{
		name:   name,
		sample: &BasicEvent{},
		// events storage
		eventFc: make(map[string]FactoryFunc),
		// listeners
		listeners:     make(map[string]*ListenerQueue),
		listenedNames: make(map[string]int),
	}

	// for async fire by goroutine
	em.ConsumerNum = 5
	em.ChanLength = 10

	// apply options
	return em.WithOptions(fns...)
}

// WithOptions create event manager with options
func (em *Manager) WithOptions(fns ...OptionFn) *Manager {
	for _, fn := range fns {
		fn(&em.Options)
	}
	return em
}

/*************************************************************
 * -- register listeners
 *************************************************************/

// AddListener alias of the method On()
func (em *Manager) AddListener(name string, listener Listener, priority ...int) {
	em.On(name, listener, priority...)
}

// Listen alias of the On()
func (em *Manager) Listen(name string, listener Listener, priority ...int) {
	em.On(name, listener, priority...)
}

// On register a event handler/listener. can setting priority.
//
// Usage:
//
//	em.On("evt0", listener)
//	em.On("evt0", listener, High)
func (em *Manager) On(name string, listener Listener, priority ...int) {
	pv := Normal
	if len(priority) > 0 {
		pv = priority[0]
	}

	em.addListenerItem(name, &ListenerItem{pv, listener})
}

// Subscribe alias of the AddSubscriber()
func (em *Manager) Subscribe(sbr Subscriber) {
	em.AddSubscriber(sbr)
}

// AddSubscriber add events by subscriber interface.
// you can register multi event listeners in a struct func.
// more usage please see README or test.
func (em *Manager) AddSubscriber(sbr Subscriber) {
	for name, listener := range sbr.SubscribedEvents() {
		switch lt := listener.(type) {
		case Listener:
			em.On(name, lt)
		// case ListenerFunc:
		// 	em.On(name, lt)
		case ListenerItem:
			em.addListenerItem(name, &lt)
		default:
			panic("event: the value must be an Listener or ListenerItem instance")
		}
	}
}

func (em *Manager) addListenerItem(name string, li *ListenerItem) {
	name = goodName(name, true)
	if li.Listener == nil {
		panicf("event: the event %q listener cannot be empty", name)
	}
	if reflect.ValueOf(li.Listener).Kind() == reflect.Struct {
		panicf("event: %q - listener should be pointer of Listener", name)
	}

	// exists, append it.
	if lq, ok := em.listeners[name]; ok {
		lq.Push(li)
	} else { // first add.
		em.listenedNames[name] = 1
		em.listeners[name] = (&ListenerQueue{}).Push(li)
	}
}

/*************************************************************
 * Listener Manage: - trigger event
 *************************************************************/

// MustTrigger alias of the method MustFire()
func (em *Manager) MustTrigger(name string, params M) Event {
	return em.MustFire(name, params)
}

// MustFire fire event by name. will panic on error
func (em *Manager) MustFire(name string, params M) Event {
	err, e := em.Fire(name, params)
	if err != nil {
		panic(err)
	}
	return e
}

// Trigger alias of the method Fire()
func (em *Manager) Trigger(name string, params M) (error, Event) {
	return em.Fire(name, params)
}

// Fire trigger event by name. if not found listener, will return (nil, nil)
func (em *Manager) Fire(name string, params M) (err error, e Event) {
	name = goodName(name, false)

	// use pre-defined Event
	if fc, ok := em.eventFc[name]; ok {
		e = fc() // make new instance
		if params != nil {
			e.SetData(params)
		}
	} else {
		// create a basic event instance
		e = em.newBasicEvent(name, params)
	}

	// call listeners handle event
	err = em.FireEvent(e)
	return
}

// FireEvent fire event by given Event instance
func (em *Manager) FireEvent(e Event) (err error) {
	if em.EnableLock {
		em.Lock()
		defer em.Unlock()
	}

	// ensure aborted is false.
	e.Abort(false)
	name := e.Name()

	// fire direct matched listeners. eg: db.user.add
	if lq, ok := em.listeners[name]; ok {
		// sort by priority before call.
		for _, li := range lq.Sort().Items() {
			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				return
			}
		}
	}

	// fire group listeners by wildcard. eg "db.user.*"
	if em.MatchMode == ModePath {
		err = em.firePathMode(name, e)
		return
	}

	// handle mode: ModeSimple
	err = em.fireSimpleMode(name, e)
	if err != nil || e.IsAborted() {
		return
	}

	// fire wildcard event listeners
	if lq, ok := em.listeners[Wildcard]; ok {
		for _, li := range lq.Sort().Items() {
			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				break
			}
		}
	}
	return
}

// ModeSimple has group listeners by wildcard. eg "db.user.*"
//
// Example:
//   - event "db.user.add" will trigger listeners on the "db.user.*"
func (em *Manager) fireSimpleMode(name string, e Event) (err error) {
	pos := strings.LastIndexByte(name, '.')

	if pos > 0 && pos < len(name) {
		groupName := name[:pos+1] + Wildcard // "app.*"

		if lq, ok := em.listeners[groupName]; ok {
			for _, li := range lq.Sort().Items() {
				err = li.Listener.Handle(e)
				if err != nil || e.IsAborted() {
					return
				}
			}
		}
	}

	return nil
}

// ModePath fire group listeners by ModePath.
//
// Example:
//   - event "db.user.add" will trigger listeners on the "db.**"
//   - event "db.user.add" will trigger listeners on the "db.user.*"
func (em *Manager) firePathMode(name string, e Event) (err error) {
	if strings.IndexByte(name, '.') < 0 {
		return nil
	}

	for pattern, lq := range em.listeners {
		if pattern != name && matchNodePath(pattern, name, ".") {
			for _, li := range lq.Sort().Items() {
				err = li.Listener.Handle(e)
				if err != nil || e.IsAborted() {
					return
				}
			}
		}
	}

	return nil
}

// Fire2 alias of the method Fire()
func (em *Manager) Fire2(name string, params M) (error, Event) {
	return em.Fire(name, params)
}

// FireByChan async fire event by go channel
func (em *Manager) FireByChan(e Event) {
	// once make
	em.once.Do(func() {
		em.makeConsumers()
	})

	// dispatch event
	em.ch <- e
}

// async fire event by 'go' keywords
func (em *Manager) makeConsumers() {
	em.ch = make(chan Event, em.ChanLength)

	// make event consumers
	for i := 0; i < em.ConsumerNum; i++ {
		go func() {
			for e := range em.ch {
				_ = em.FireEvent(e) // ignore async fire error
			}
		}()
	}
}

// FireBatch fire multi event at once.
//
// Usage:
//
//	FireBatch("name1", "name2", &MyEvent{})
func (em *Manager) FireBatch(es ...any) (ers []error) {
	var err error
	for _, e := range es {
		if name, ok := e.(string); ok {
			err, _ = em.Fire(name, nil)
		} else if evt, ok := e.(Event); ok {
			err = em.FireEvent(evt)
		} // ignore invalid param.

		if err != nil {
			ers = append(ers, err)
		}
	}
	return
}

// AsyncFire async fire event by 'go' keywords
func (em *Manager) AsyncFire(e Event) {
	go func(e Event) {
		_ = em.FireEvent(e)
	}(e)
}

// AwaitFire async fire event by 'go' keywords, but will wait return result
func (em *Manager) AwaitFire(e Event) (err error) {
	ch := make(chan error)

	go func(e Event) {
		err := em.FireEvent(e)
		ch <- err
	}(e)

	err = <-ch
	close(ch)
	return
}

/*************************************************************
 * Event Manage
 *************************************************************/

// AddEvent add a pre-defined event instance to manager.
func (em *Manager) AddEvent(e Event) {
	name := goodName(e.Name(), false)

	if ec, ok := e.(Cloneable); ok {
		em.AddEventFc(name, func() Event {
			return ec.Clone()
		})
	} else {
		em.AddEventFc(name, func() Event {
			return e
		})
	}
}

// AddEventFc add a pre-defined event factory func to manager.
func (em *Manager) AddEventFc(name string, fc FactoryFunc) {
	em.eventFc[name] = fc
}

// GetEvent get a pre-defined event instance by name
func (em *Manager) GetEvent(name string) (e Event, ok bool) {
	fc, ok := em.eventFc[name]
	if ok {
		return fc(), true
	}
	return
}

// HasEvent has pre-defined event check
func (em *Manager) HasEvent(name string) bool {
	_, ok := em.eventFc[name]
	return ok
}

// RemoveEvent delete pre-define Event by name
func (em *Manager) RemoveEvent(name string) {
	if _, ok := em.eventFc[name]; ok {
		delete(em.eventFc, name)
	}
}

// RemoveEvents remove all registered events
func (em *Manager) RemoveEvents() {
	em.eventFc = map[string]FactoryFunc{}
}

/*************************************************************
 * Helper Methods
 *************************************************************/

// newBasicEvent create new BasicEvent by clone em.sample
func (em *Manager) newBasicEvent(name string, data M) *BasicEvent {
	var cp = *em.sample

	cp.SetName(name)
	cp.SetData(data)
	return &cp
}

// HasListeners check has direct listeners for the event name.
func (em *Manager) HasListeners(name string) bool {
	_, ok := em.listenedNames[name]
	return ok
}

// Listeners get all listeners
func (em *Manager) Listeners() map[string]*ListenerQueue {
	return em.listeners
}

// ListenersByName get listeners by given event name
func (em *Manager) ListenersByName(name string) *ListenerQueue {
	return em.listeners[name]
}

// ListenersCount get listeners number for the event name.
func (em *Manager) ListenersCount(name string) int {
	if lq, ok := em.listeners[name]; ok {
		return lq.Len()
	}
	return 0
}

// ListenedNames get listened event names
func (em *Manager) ListenedNames() map[string]int {
	return em.listenedNames
}

// RemoveListener remove a given listener, you can limit event name.
//
// Usage:
//
//	RemoveListener("", listener)
//	RemoveListener("name", listener) // limit event name.
func (em *Manager) RemoveListener(name string, listener Listener) {
	if name != "" {
		if lq, ok := em.listeners[name]; ok {
			lq.Remove(listener)

			// delete from manager
			if lq.IsEmpty() {
				delete(em.listeners, name)
				delete(em.listenedNames, name)
			}
		}
		return
	}

	// name is empty. find all listener and remove matched.
	for name, lq := range em.listeners {
		lq.Remove(listener)

		// delete from manager
		if lq.IsEmpty() {
			delete(em.listeners, name)
			delete(em.listenedNames, name)
		}
	}
}

// RemoveListeners remove listeners by given name
func (em *Manager) RemoveListeners(name string) {
	_, ok := em.listenedNames[name]
	if ok {
		em.listeners[name].Clear()

		// delete from manager
		delete(em.listeners, name)
		delete(em.listenedNames, name)
	}
}

// Clear alias of the Reset()
func (em *Manager) Clear() {
	em.Reset()
}

// Reset the manager, clear all data.
func (em *Manager) Reset() {
	// clear all listeners
	for _, lq := range em.listeners {
		lq.Clear()
	}

	// reset all
	em.name = ""
	em.eventFc = make(map[string]FactoryFunc)
	em.listeners = make(map[string]*ListenerQueue)
	em.listenedNames = make(map[string]int)
}
