// Package event is lightweight event manager and dispatcher implements by Go.
//
package event

// Event interface
type Event interface {
	Name() string
	Data() []interface{}
	Abort()
	Aborted() bool
}

// BasicEvent a basic event struct define.
type BasicEvent struct {
	// event name
	name string
	// user data.
	data []interface{}
	// target
	target  interface{}
	aborted bool
}

// NewBasic new an basic event instance
func NewBasic(name string) *BasicEvent {
	return &BasicEvent{name: name}
}

// Abort abort event exec
func (e *BasicEvent) Abort() {
	e.aborted = true
}

// Aborted check.
func (e *BasicEvent) Aborted() bool {
	return e.aborted
}

// Init event data
func (e *BasicEvent) Init(name string, target interface{}, data ...interface{}) *BasicEvent {
	e.SetName(name)
	e.Fill(target, data...)
	return e
}

// Fill event data
func (e *BasicEvent) Fill(target interface{}, data ...interface{}) *BasicEvent {
	e.data = data
	e.target = target
	return e
}

func (e *BasicEvent) reset() {
	e.name = ""
	e.data = make([]interface{}, 0)
	e.target = nil
	e.aborted = false
}

// Get get data by index
func (e *BasicEvent) Get(index int) interface{} {
	if len(e.data) > index {
		return e.data[index]
	}

	return nil
}

// Name get name
func (e *BasicEvent) Name() string {
	return e.name
}

// Data get all data
func (e *BasicEvent) Data() []interface{} {
	return e.data
}

// SetData set data to the event
func (e *BasicEvent) SetData(data ...interface{}) {
	e.data = data
}

// SetName set name
func (e *BasicEvent) SetName(name string) {
	e.name = name
}

// SetTarget set target
func (e *BasicEvent) SetTarget(target interface{}) {
	e.target = target
}

// Target get target
func (e *BasicEvent) Target() interface{} {
	return e.target
}
