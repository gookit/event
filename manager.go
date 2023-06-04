package event

import (
	"reflect"
	"strings"
	"sync"
)

// Wildcard event name
const Wildcard = "*"
const AnyNode = "*"
const AllNode = "**"

const (
	// ModeSimple old mode, "*" will match all to end.
	ModeSimple uint8 = iota
	// ModePath path mode.
	//  - "*" matches any sequence of non-/ characters (like at path.Match())
	//  - "**" match all characters to end.
	ModePath
)

// M is short name for map[string]...
type M = map[string]any

// ManagerFace event manager interface
type ManagerFace interface {
	// AddEvent events: add event
	AddEvent(Event)
	// On listeners: add listeners
	On(name string, listener Listener, priority ...int)
	// Fire event
	Fire(name string, params M) (error, Event)
}

// Manager event manager definition. for manage events and listeners
type Manager struct {
	sync.Mutex
	// EnableLock enable lock on fire event.
	EnableLock bool
	// ChanLength for fire events by goroutine
	ChanLength  int
	ConsumerNum int
	// MatchMode match mode. default is ModeSimple
	MatchMode uint8

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
func NewManager(name string) *Manager {
	em := &Manager{
		name:   name,
		sample: &BasicEvent{},
		// events storage
		eventFc: make(map[string]FactoryFunc),
		// listeners
		listeners:     make(map[string]*ListenerQueue),
		listenedNames: make(map[string]int),
		// for async fire by goroutine
		ChanLength:  10,
		ConsumerNum: 5,
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
	name, _ = goodName(name)
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
	var fireAll bool
	name, fireAll = goodName(name)
	if fireAll {
		return em.fireAll(name, params)
	}

	// NOTICE: must check the '*' global listeners
	if false == em.HasListeners(name) && false == em.HasListeners(Wildcard) {
		// has group listeners. "app.*" "app.db.*"
		// eg: "app.db.run" will trigger listeners on the "app.db.*"
		pos := strings.LastIndexByte(name, '.')
		if pos < 0 || pos == len(name)-1 {
			return // not found listeners.
		}

		groupName := name[:pos+1] + Wildcard // "app.db.*"
		if false == em.HasListeners(groupName) {
			return // not found listeners.
		}
	}

	// call listeners use defined Event
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

func (em *Manager) definedEvent(e Event) Event {
	e1 := e

	return e1
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

	// find matched listeners
	lq, ok := em.listeners[name]
	if ok {
		// sort by priority before call.
		for _, li := range lq.Sort().Items() {
			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				return
			}
		}
	}

	// has group listeners. "app.*" "app.db.*"
	// eg: "app.run" will trigger listeners on the "app.*"
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

	// has wildcard event listeners
	if lq, ok := em.listeners[AllNode]; ok {
		for _, li := range lq.Sort().Items() {
			err = li.Listener.Handle(e)
			if err != nil || e.IsAborted() {
				break
			}
		}
	}
	return
}

// matchListeners has listeners for the event name.
func (em *Manager) matchListeners(name string) []string {
	_, ok := em.listenedNames[name]
	return ok
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
	name, _ := goodName(e.Name())

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
	name, _ = goodName(name)
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

// HasListeners has listeners for the event name.
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
