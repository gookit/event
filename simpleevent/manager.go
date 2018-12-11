package simpleevent

import (
	"strings"
	"sync"
)

const wildcard = "*"

// EventHandler func define
type EventHandler func(e *EventData) error

/*************************************************************
 * Event Manager
 *************************************************************/

// EventManager struct
type EventManager struct {
	names  map[string]int
	events map[string][]EventHandler
	pool   sync.Pool
}

// NewEventManager create EventManager instance
func NewEventManager() *EventManager {
	em := &EventManager{
		names:  make(map[string]int),
		events: make(map[string][]EventHandler),
	}

	// set pool creator
	em.pool.New = func() interface{} {
		return &EventData{}
	}

	return em
}

// On register a event handler
func (em *EventManager) On(name string, handler EventHandler) {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("event name cannot be empty")
	}

	if ls, ok := em.events[name]; ok {
		em.events[name] = append(ls, handler)
	} else { // first add.
		em.events[name] = []EventHandler{handler}
	}
}

// MustFire fire event by name
func (em *EventManager) MustFire(name string, args ...interface{}) {
	err := em.Fire(name, args...)
	if err != nil {
		panic(err)
	}
}

// Fire event by name
func (em *EventManager) Fire(name string, args ...interface{}) (err error) {
	handlers, ok := em.events[name]
	if !ok {
		return
	}

	e := em.pool.Get().(*EventData)
	e.init(name, args)

	defer func() {
		e.reset()
		em.pool.Put(e)
	}()

	err = em.callHandlers(e, handlers)
	if err != nil || e.Aborted() {
		return
	}

	// group listen "app.*"
	// groupName :=

	// wildcard event handler
	if em.HasEvent(wildcard) {
		err = em.callHandlers(e, em.events[wildcard])
	}

	return
}

func (em *EventManager) callHandlers(e *EventData, handlers []EventHandler) (err error) {
	for _, handler := range handlers {
		err = handler(e)
		if err != nil || e.Aborted() {
			return
		}
	}
	return
}

// HasEvent has event check
func (em *EventManager) HasEvent(name string) bool {
	_, ok := em.names[name]
	return ok
}

// DelEventsByName delete events by name
func (em *EventManager) DelEventsByName(name string) bool {
	_, ok := em.names[name]
	if ok {
		delete(em.names, name)
		delete(em.events, name)
	}
	return ok
}

// EventsByName get events and handlers by name
func (em *EventManager) EventsByName(name string) (es []EventHandler) {
	es, _ = em.events[name]
	return
}

// Events get all events and handlers
func (em *EventManager) Events() map[string][]EventHandler {
	return em.events
}

// Names get all event names
func (em *EventManager) Names() map[string]int {
	return em.names
}

// Clear all events info.
func (em *EventManager) Clear() {
	em.events = map[string][]EventHandler{}
}
