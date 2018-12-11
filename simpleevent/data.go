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
	Data []interface{}
}

// Name get
func (e *EventData) Name() string {
	return e.name
}

// Abort abort event exec
func (e *EventData) Abort() {
	e.aborted = true
}

// Aborted check.
func (e *EventData) Aborted() bool {
	return e.aborted
}

func (e *EventData) init(s string, args []interface{}) {
	e.name = s
	e.Data = args
}

func (e *EventData) reset() {
	e.name = ""
	e.Data = make([]interface{}, 0)
}
