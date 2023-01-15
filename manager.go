package event

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// Wildcard event name
const Wildcard = "*"

// regex for check good event name.
var goodNameReg = regexp.MustCompile(`^[a-zA-Z][\w-.*]*$`)

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
	// name of the manager
	name string
	// pool sync.Pool
	// is a sample for new BasicEvent
	sample *BasicEvent
	// storage user custom Event instance. you can pre-define some Event instances.
	events map[string]Event
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
		events: make(map[string]Event),
		// listeners
		listeners:     make(map[string]*ListenerQueue),
		listenedNames: make(map[string]int),
	}

	return em
}

/*************************************************************
 * Listener Manage: - register listener
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
//	On("evt0", listener)
//	On("evt0", listener, High)
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
	if name != Wildcard {
		name = goodName(name)
	}

	if li.Listener == nil {
		panic("event: the event '" + name + "' listener cannot be empty")
	}
	if reflect.ValueOf(li.Listener).Kind() == reflect.Struct {
		panic("don't use struct Listener, can be pointer Listener")
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
	name = goodName(name)

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
	if e, ok := em.events[name]; ok {
		if params != nil {
			e.SetData(params)
		}

		err = em.FireEvent(e)
		return err, e
	}

	// create a basic event instance
	e = em.newBasicEvent(name, params)
	// call listeners handle event
	err = em.FireEvent(e)
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

// FireBatch fire multi event at once.
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

/*************************************************************
 * Event Manage
 *************************************************************/

// AddEvent add a defined event instance to manager.
func (em *Manager) AddEvent(e Event) {
	name := goodName(e.Name())
	em.events[name] = e
}

// GetEvent get a defined event instance by name
func (em *Manager) GetEvent(name string) (e Event, ok bool) {
	e, ok = em.events[name]
	return
}

// HasEvent has event check
func (em *Manager) HasEvent(name string) bool {
	_, ok := em.events[name]
	return ok
}

// RemoveEvent delete Event by name
func (em *Manager) RemoveEvent(name string) {
	if _, ok := em.events[name]; ok {
		delete(em.events, name)
	}
}

// RemoveEvents remove all registered events
func (em *Manager) RemoveEvents() {
	em.events = map[string]Event{}
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
	em.events = make(map[string]Event)
	em.listeners = make(map[string]*ListenerQueue)
	em.listenedNames = make(map[string]int)
}

func goodName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("event: the event name cannot be empty")
	}

	if !goodNameReg.MatchString(name) {
		panic(`event: the event name is invalid, must match regex '^[a-zA-Z][\w-.]*$'`)
	}

	return name
}
