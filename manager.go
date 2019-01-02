package event

import (
	"regexp"
	"strings"
	"sync"
)

// Wildcard event name
const Wildcard = "*"

// M is short name fo map[string]...
type M map[string]interface{}

// ManagerFace event manager interface
type ManagerFace interface {
	// events
	AddEvent(Event)
	// listeners
	On(name string, listener Listener, priority ...int)
	Fire(name string, params M) error
}

// Manager event manager definition. for manage events and listeners
type Manager struct {
	name string
	pool sync.Pool
	// storage user custom Event instance. you can pre-define some Event instances.
	events map[string]Event
	// storage all event name and ListenerQueue map
	listeners map[string]*ListenerQueue
	// storage all event names by listened
	listenedNames map[string]int
}

var goodNameReg = regexp.MustCompile(`^[a-zA-Z][\w-.*]*$`)

// NewManager create event manager
func NewManager(name string) *Manager {
	em := &Manager{
		name:          name,
		events:        make(map[string]Event),
		listeners:     make(map[string]*ListenerQueue),
		listenedNames: make(map[string]int),
	}

	// set pool creator
	em.pool.New = func() interface{} {
		return &BasicEvent{}
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

// On register a event handler/listener. can setting priority.
// Usage:
// 	On("evt0", listener)
// 	On("evt0", listener, High)
func (em *Manager) On(name string, listener Listener, priority ...int) {
	pv := Normal
	if len(priority) > 0 {
		pv = priority[0]
	}

	em.addListenerItem(name, &ListenerItem{pv, listener})
}

// AddSubscriber add event subscriber.
// you can register multi event listeners in a struct func.
// more usage please see README or test.
func (em *Manager) AddSubscriber(sbr Subscriber) {
	for name, listener := range sbr.SubscribeEvents() {
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

// MustFire fire event by name. will panic on error
func (em *Manager) MustFire(name string, params M) {
	err := em.Fire(name, params)
	if err != nil {
		panic(err)
	}
}

// Trigger alias of the method Fire()
func (em *Manager) Trigger(name string, params M) (err error) {
	return em.Fire(name, params)
}

// Fire trigger event by name
func (em *Manager) Fire(name string, params M) (err error) {
	name = goodName(name)

	// call listeners use defined Event
	if e, ok := em.events[name]; ok {
		if params != nil {
			e.SetData(params)
		}
		return em.FireEvent(e)
	}

	// create a basic event instance
	e := em.pool.Get().(*BasicEvent)
	e.SetName(name)
	e.SetData(params)

	// call listeners
	err = em.FireEvent(e)

	e.reset()
	em.pool.Put(e)
	return
}

// AsyncFire async fire event by 'go' keywords
func (em *Manager) AsyncFire(e Event) {
	go func(e Event) {
		err := em.FireEvent(e)
		if err != nil {
			panic(err)
		}
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
	return
}

// FireBatch fire multi event at once.
// Usage:
//	FireBatch("name1", "name2", &MyEvent{})
func (em *Manager) FireBatch(es ...interface{}) (ers []error) {
	var err error
	for _, e := range es {
		if name, ok := e.(string); ok {
			err = em.Fire(name, nil)
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
	// find matched listeners
	name := e.Name()
	lq, ok := em.listeners[name]
	if !ok {
		return
	}

	// ensure aborted is false.
	e.Abort(false)

	// sort by priority before call.
	for _, li := range lq.Sort().Items() {
		err = li.Listener.Handle(e)
		if err != nil || e.IsAborted() {
			return
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
				return
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

// HasListeners has listeners for the event name.
func (em *Manager) HasListeners(name string) bool {
	_, ok := em.listeners[name]
	return ok
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
// Usage:
// 	RemoveListener("", listener)
// 	RemoveListener("name", listener) // limit event name.
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

// Clear all data
func (em *Manager) Clear() {
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
