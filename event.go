// Package event is lightweight event manager and dispatcher implements by Go.
package event

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

// BasicEvent a basic event struct define.
type BasicEvent struct {
	// event name
	name string
	// user data.
	data map[string]any
	// target
	target any
	// mark is aborted
	aborted bool
}

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
func (e *BasicEvent) Abort(abort bool) {
	e.aborted = abort
}

// Fill event data
func (e *BasicEvent) Fill(target any, data M) *BasicEvent {
	if data != nil {
		e.data = data
	}

	e.target = target
	return e
}

// AttachTo add current event to the event manager.
func (e *BasicEvent) AttachTo(em ManagerFace) {
	em.AddEvent(e)
}

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
func (e *BasicEvent) Name() string {
	return e.name
}

// Data get all data
func (e *BasicEvent) Data() map[string]any {
	return e.data
}

// IsAborted check.
func (e *BasicEvent) IsAborted() bool {
	return e.aborted
}

// Target get target
func (e *BasicEvent) Target() any {
	return e.target
}

// SetName set event name
func (e *BasicEvent) SetName(name string) *BasicEvent {
	e.name = name
	return e
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
