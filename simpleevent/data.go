package simpleevent

/*************************************************************
 * Event Data
 *************************************************************/

// EventData struct
type EventData struct {
	aborted bool
	// event name
	name string
	// user data.
	Data []any
}

// Name get
func (e *EventData) Name() string {
	return e.name
}

// Abort abort event exec
func (e *EventData) Abort() {
	e.aborted = true
}

// IsAborted check.
func (e *EventData) IsAborted() bool {
	return e.aborted
}

func (e *EventData) init(name string, data []any) {
	e.name = name
	e.Data = data
}

func (e *EventData) reset() {
	e.name = ""
	e.Data = make([]any, 0)
	e.aborted = false
}
