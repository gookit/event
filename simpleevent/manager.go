package simpleevent

import (
	"bytes"
	"strings"
	"sync"
)

// Wildcard event name
const Wildcard = "*"

// HandlerFunc event handler func define
type HandlerFunc func(e *EventData) error

/*************************************************************
 * Event Manager
 *************************************************************/

// EventManager struct
type EventManager struct {
	pool  sync.Pool
	names map[string]int
	// storage event handlers
	handlers map[string][]HandlerFunc
}

// NewEventManager create EventManager instance
func NewEventManager() *EventManager {
	em := &EventManager{
		names:    make(map[string]int),
		handlers: make(map[string][]HandlerFunc),
	}

	// set pool creator
	em.pool.New = func() any {
		return &EventData{}
	}

	return em
}

// On register a event handler
func (em *EventManager) On(name string, handler HandlerFunc) {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("event name cannot be empty")
	}

	if ls, ok := em.handlers[name]; ok {
		em.names[name]++
		em.handlers[name] = append(ls, handler)
	} else { // first add.
		em.names[name] = 1
		em.handlers[name] = []HandlerFunc{handler}
	}
}

// MustFire fire handlers by name. will panic on error
func (em *EventManager) MustFire(name string, args ...any) {
	err := em.Fire(name, args...)
	if err != nil {
		panic(err)
	}
}

// Fire handlers by name
func (em *EventManager) Fire(name string, args ...any) (err error) {
	handlers, ok := em.handlers[name]
	if !ok {
		return
	}

	e := em.pool.Get().(*EventData)
	e.init(name, args)

	// call event handlers
	err = em.doFire(e, handlers)

	e.reset()
	em.pool.Put(e)
	return
}

func (em *EventManager) doFire(e *EventData, handlers []HandlerFunc) (err error) {
	err = em.callHandlers(e, handlers)
	if err != nil || e.IsAborted() {
		return
	}

	// group listen "app.*"
	// groupName :=

	// Wildcard event handler
	if em.HasEvent(Wildcard) {
		err = em.callHandlers(e, em.handlers[Wildcard])
	}

	return
}

func (em *EventManager) callHandlers(e *EventData, handlers []HandlerFunc) (err error) {
	for _, handler := range handlers {
		err = handler(e)
		if err != nil || e.IsAborted() {
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

// GetEventHandlers get handlers and handlers by name
func (em *EventManager) GetEventHandlers(name string) (es []HandlerFunc) {
	es, _ = em.handlers[name]
	return
}

// EventHandlers get all event handlers
func (em *EventManager) EventHandlers() map[string][]HandlerFunc {
	return em.handlers
}

// EventNames get all event names
func (em *EventManager) EventNames() map[string]int {
	return em.names
}

// String convert to string.
func (em *EventManager) String() string {
	buf := new(bytes.Buffer)
	for name, hs := range em.handlers {
		buf.WriteString(name + " handlers:\n ")
		for _, h := range hs {
			buf.WriteString(funcName(h))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// ClearHandlers clear handlers by name
func (em *EventManager) ClearHandlers(name string) bool {
	_, ok := em.names[name]
	if ok {
		delete(em.names, name)
		delete(em.handlers, name)
	}
	return ok
}

// Clear all handlers info.
func (em *EventManager) Clear() {
	em.names = map[string]int{}
	em.handlers = map[string][]HandlerFunc{}
}
