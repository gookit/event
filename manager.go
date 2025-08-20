package event

import (
	"context"
	"reflect"
	"sync"
)

const (
	defaultChannelSize = 100
	defaultConsumerNum = 3
)

// Manager event manager definition. for manage events and listeners
type Manager struct {
	Options
	sync.Mutex

	wg  sync.WaitGroup
	ch  chan Event
	oc  sync.Once
	err error // latest error

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

// NewM create event manager. alias of the NewManager()
func NewM(name string, fns ...OptionFn) *Manager {
	return NewManager(name, fns...)
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

	// em.EnableLock = true
	// for async fire by goroutine
	em.ConsumerNum = defaultConsumerNum
	em.ChannelSize = defaultChannelSize

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
 * region Register listeners
 *************************************************************/

// AddListener register a event handler/listener. alias of the method On()
func (em *Manager) AddListener(name string, listener Listener, priority ...int) {
	em.On(name, listener, priority...)
}

// Listen register a event handler/listener. alias of the On()
func (em *Manager) Listen(name string, listener Listener, priority ...int) {
	em.On(name, listener, priority...)
}

// Listen register a event handler/listener. trigger once.
func (em *Manager) Once(name string, listener Listener, priority ...int) {
	var listenerOnce Listener
	listenerOnce = ListenerFunc(func(e Event) error {
		em.RemoveListener(name, listenerOnce)
		return listener.Handle(e)
	})
	em.On(name, listenerOnce, priority...)
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

// Subscribe add events by subscriber interface. alias of the AddSubscriber()
func (em *Manager) Subscribe(sbr Subscriber) {
	em.AddSubscriber(sbr)
}

// AddSubscriber add events by subscriber interface.
//
// you can register multi event listeners in a struct func.
// more usage please see README or tests.
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
		panicf("event: %q - struct listener must be pointer", name)
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
 * region Event Manage
 *************************************************************/

// AddEvent add a pre-defined event instance to manager.
func (em *Manager) AddEvent(e Event) error {
	name, err := goodNameOrErr(e.Name(), false)
	if err != nil {
		return err
	}

	if ec, ok := e.(Cloneable); ok {
		em.AddEventFc(name, func() Event {
			return ec.Clone()
		})
	} else {
		em.AddEventFc(name, func() Event {
			return e
		})
	}
	return nil
}

// AddEventFc add a pre-defined event factory func to manager.
func (em *Manager) AddEventFc(name string, fc FactoryFunc) {
	em.Lock()
	em.eventFc[name] = fc
	em.Unlock()
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
 * region Helper Methods
 *************************************************************/

// newBasicEvent create new BasicEvent by clone em.sample
func (em *Manager) newBasicEvent(name string, data M) *BasicEvent {
	var cp = *em.sample
	cp.ctx = context.Background()

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
func (em *Manager) Clear() { em.Reset() }

// Reset the manager, clear all data.
func (em *Manager) Reset() {
	// clear all listeners
	for _, lq := range em.listeners {
		lq.Clear()
	}

	// reset all
	em.ch = nil
	em.oc = sync.Once{}
	em.wg = sync.WaitGroup{}

	em.eventFc = make(map[string]FactoryFunc)
	em.listeners = make(map[string]*ListenerQueue)
	em.listenedNames = make(map[string]int)
}
