// Package event is lightweight event manager and dispatcher implements by Go.
package event

import (
	"context"
)

// wildcard event name
const (
	Wildcard = "*"
	AnyNode  = "*"
	AllNode  = "**"
)

const (
	// ModeSimple old mode, simple match group listener.
	//
	//  - "*" only allow one and must at end
	//
	// Support: "user.*" -> match "user.created" "user.updated"
	ModeSimple uint8 = iota

	// ModePath path mode.
	//
	//  - "*" matches any sequence of non . characters (like at path.Match())
	//  - "**" match all characters to end, only allow at start or end on pattern.
	//
	// Support like this:
	// 	"eve.some.*.*"       -> match "eve.some.thing.run" "eve.some.thing.do"
	// 	"eve.some.*.run"     -> match "eve.some.thing.run", but not match "eve.some.thing.do"
	// 	"eve.some.**"        -> match any start with "eve.some.". eg: "eve.some.thing.run" "eve.some.thing.do"
	// 	"**.thing.run"       -> match any ends with ".thing.run". eg: "eve.some.thing.run"
	ModePath
)

// M is short name for map[string]...
type M = map[string]any

// ManagerFace event manager interface
type ManagerFace interface {
	// AddEvent events: add event
	AddEvent(Event) error
	// On listeners: add listeners
	On(name string, listener Listener, priority ...int)
	// Fire event
	Fire(name string, params M) (error, Event)
}

// Options event manager config options
type Options struct {
	// EnableLock enable lock on fire event. default is False.
	EnableLock bool
	// ChannelSize for fire events by goroutine
	ChannelSize int
	ConsumerNum int
	// MatchMode event name match mode. default is ModeSimple
	MatchMode uint8
}

// OptionFn event manager config option func
type OptionFn func(o *Options)

// UsePathMode set event name match mode to ModePath
func UsePathMode(o *Options) { o.MatchMode = ModePath }

// EnableLock enable lock on fire event.
func EnableLock(enable bool) OptionFn {
	return func(o *Options) {
		o.EnableLock = enable
	}
}

// Event interface
type Event interface {
	Name() string
	Get(key string) any
	Set(key string, val any)
	Add(key string, val any)
	Data() map[string]any
	SetData(M) Event
	Abort(bool)
	IsAborted() bool
}

// Cloneable interface. event can be cloned.
//
// Check and convert:
//	if ec, ok := e.(Cloneable); ok {}
type Cloneable interface {
	Event
	Clone() Event
}

// ContextAble context-able event interface
//
// Check and convert in listener:
//	if ec, ok := e.(ContextAble); ok {}
type ContextAble interface {
	Event
	Context() context.Context
	WithContext(ctx context.Context)
}

// FactoryFunc for create event instance.
type FactoryFunc func() Event

// BasicEvent a built-in implements Event interface
type BasicEvent struct {
	ContextTrait
	// event name
	name string
	// user data.
	data map[string]any
	// target
	target any
	// mark is aborted
	aborted bool
}

// New create an event instance
func New(name string, data M) *BasicEvent { return NewBasic(name, data) }

// NewEvent create an event instance
func NewEvent(name string, data M) *BasicEvent { return NewBasic(name, data) }

// NewBasic new a basic event instance
func NewBasic(name string, data M) *BasicEvent {
	if data == nil {
		data = make(map[string]any)
	}

	return &BasicEvent{
		name: name,
		data: data,
	}
}

// Abort event loop exec
func (e *BasicEvent) Abort(abort bool) { e.aborted = abort }

// Fill event data
func (e *BasicEvent) Fill(target any, data M) *BasicEvent {
	if data != nil {
		e.data = data
	}

	e.target = target
	return e
}

// AttachTo add current event to the event manager.
func (e *BasicEvent) AttachTo(em ManagerFace) error { return em.AddEvent(e) }

// Get data by index
func (e *BasicEvent) Get(key string) any {
	if v, ok := e.data[key]; ok {
		return v
	}
	return nil
}

// Add value by key
func (e *BasicEvent) Add(key string, val any) {
	if _, ok := e.data[key]; !ok {
		e.Set(key, val)
	}
}

// Set value by key
func (e *BasicEvent) Set(key string, val any) {
	if e.data == nil {
		e.data = make(map[string]any)
	}
	e.data[key] = val
}

// Name get event name
func (e *BasicEvent) Name() string { return e.name }

// Data get all data
func (e *BasicEvent) Data() map[string]any { return e.data }

// IsAborted check.
func (e *BasicEvent) IsAborted() bool { return e.aborted }

// Target get target
func (e *BasicEvent) Target() any { return e.target }

// SetName set event name
func (e *BasicEvent) SetName(name string) *BasicEvent {
	e.name = name
	return e
}

// Clone new instance
func (e *BasicEvent) Clone() Event {
	var cp = *e
	return &cp
}

// SetData set data to the event
func (e *BasicEvent) SetData(data M) Event {
	if data != nil {
		e.data = data
	}
	return e
}

// SetTarget set event target
func (e *BasicEvent) SetTarget(target any) *BasicEvent {
	e.target = target
	return e
}

// ContextTrait event context trait
type ContextTrait struct {
	// context
	ctx context.Context
}

// Context get context
func (t *ContextTrait) Context() context.Context {
	if t.ctx == nil {
		return context.Background()
	}
	return t.ctx
}

// WithContext set context
func (t *ContextTrait) WithContext(ctx context.Context) {
	t.ctx = ctx
}