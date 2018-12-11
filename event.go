package event

// EventFace
type EventFace interface {
	Params()
	Name() string
	Stop(bool)
}

// Event
type Event struct {
	name    string
	stopped bool
	target  interface{}
	params  interface{}
}

// New event instance
func New(name string) *Event {
	return &Event{name: name}
}

// Init
func (e *Event) Init(target, params interface{}) *Event {
	return e
}

// Clone current event object
func (e *Event) Clone() *Event {
	ne := *e
	ne.name = ""
	ne.target = nil
	ne.params = nil
	ne.stopped = false
	return &ne
}

func (e *Event) Params() interface{} {
	return e.params
}

func (e *Event) SetParams(params interface{}) {
	e.params = params
}

func (e *Event) Name() string {
	return e.name
}

func (e *Event) SetName(name string) {
	e.name = name
}

func (e *Event) SetTarget(target interface{}) {
	e.target = target
}

func (e *Event) Target() interface{} {
	return e.target
}

// Stop
func (e *Event) Stop(stopped bool) {
	e.stopped = stopped
}
