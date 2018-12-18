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
 * Listener Manage
 *************************************************************/

// On register a event handler/listener. can setting priority.
// Usage:
// 	On("evt0", listener)
// 	On("evt0", listener, High)
func (em *Manager) On(name string, listener Listener, priority ...int) {
	name = goodName(name)

	if listener == nil {
		panic("event: the event '" + name + "' listener cannot be empty")
	}

	pv := Normal
	if len(priority) > 0 {
		pv = priority[0]
	}

	li := &ListenerItem{pv, listener}

	if lq, ok := em.listeners[name]; ok {
		lq.Push(li) // append.
		em.listenedNames[name]++
	} else { // first add.
		em.listenedNames[name] = 1
		em.listeners[name] = (&ListenerQueue{}).Push(li)
	}
}

// MustFire fire event by name. will panic on error
func (em *Manager) MustFire(name string, params M) {
	err := em.Fire(name, params)
	if err != nil {
		panic(err)
	}
}

// Fire event by name
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
// TODO ... 实验性的
func (em *Manager) AsyncFire(e Event) (err error) {
	go func(e Event) {
		_ = em.FireEvent(e)
	}(e)

	return nil
}

// FireEvent fire event by given BasicEvent instance
func (em *Manager) FireEvent(e Event) (err error) {
	// find matched listeners
	name := e.Name()
	lq, ok := em.listeners[name]
	if !ok {
		return
	}

	// sort by priority before call.
	for _, li := range lq.Sort().Items() {
		err = li.listener.Handle(e)
		if err != nil || e.IsAborted() {
			return
		}
	}

	// has group listeners. "app.*" "app.db.*"
	// eg: "app.run" will trigger listeners on the "app.*"
	pos := strings.LastIndexByte(name, '.')
	if pos > 0 && pos < len(name) {
		groupName := name[:pos] + Wildcard // "app.*"

		if lq, ok := em.listeners[groupName]; ok {
			for _, li := range lq.Sort().Items() {
				err = li.listener.Handle(e)
				if err != nil || e.IsAborted() {
					return
				}
			}
		}
	}

	// has wildcard event listeners
	if lq, ok := em.listeners[Wildcard]; ok {
		for _, li := range lq.Sort().Items() {
			err = li.listener.Handle(e)
			if err != nil || e.IsAborted() {
				return
			}
		}
	}
	return
}

// HasListeners has listeners for the event name.
func (em *Manager) HasListeners(name string) bool {
	_, ok := em.listenedNames[name]
	return ok
}

// ListenersCount get listeners number for the event name.
func (em *Manager) ListenersCount(name string) int {
	if lq, ok := em.listeners[name]; ok {
		return lq.Len()
	}
	return 0
}

// ClearListeners by name
func (em *Manager) ClearListeners(name string) {
	_, ok := em.listenedNames[name]
	if ok {
		em.listeners[name].Clear()

		// delete from manager
		delete(em.listeners, name)
		delete(em.listenedNames, name)
	}
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

// DelEvent delete Event by name
func (em *Manager) DelEvent(name string) {
	if _, ok := em.events[name]; ok {
		delete(em.events, name)
	}
}

// ClearEvents clear all events
func (em *Manager) ClearEvents() {
	em.events = map[string]Event{}
}

/*************************************************************
 * Helper Methods
 *************************************************************/

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
