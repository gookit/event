package event

// DefaultEM default event manager
var DefaultEM = NewManager("default")

/*************************************************************
 * Listener
 *************************************************************/

// On register a listener to the event
func On(name string, listener Listener, priority ...int) {
	DefaultEM.On(name, listener, priority...)
}

// Listen register a listener to the event
func Listen(name string, listener Listener, priority ...int) {
	DefaultEM.Listen(name, listener, priority...)
}

// Subscribe register a listener to the event
func Subscribe(sbr Subscriber) {
	DefaultEM.Subscribe(sbr)
}

// AddSubscriber register a listener to the event
func AddSubscriber(sbr Subscriber) {
	DefaultEM.AddSubscriber(sbr)
}

// AsyncFire async fire event by 'go' keywords
func AsyncFire(e Event) {
	DefaultEM.AsyncFire(e)
}

// Trigger alias of Fire
func Trigger(name string, params M) (error, Event) {
	return DefaultEM.Fire(name, params)
}

// Fire fire listeners by name.
func Fire(name string, params M) (error, Event) {
	return DefaultEM.Fire(name, params)
}

// FireEvent fire listeners by Event instance.
func FireEvent(e Event) error {
	return DefaultEM.FireEvent(e)
}

// TriggerEvent alias of FireEvent
func TriggerEvent(e Event) error {
	return DefaultEM.FireEvent(e)
}

// MustFire fire event by name. will panic on error
func MustFire(name string, params M) Event {
	return DefaultEM.MustFire(name, params)
}

// MustTrigger alias of MustFire
func MustTrigger(name string, params M) Event {
	return DefaultEM.MustFire(name, params)
}

// FireBatch fire multi event at once.
func FireBatch(es ...interface{}) []error {
	return DefaultEM.FireBatch(es...)
}

// HasListeners has listeners for the event name.
func HasListeners(name string) bool {
	return DefaultEM.HasListeners(name)
}

// Reset the default event manager
func Reset() {
	DefaultEM.Clear()
}

/*************************************************************
 * Event
 *************************************************************/

// AddEvent has event check.
func AddEvent(e Event) {
	DefaultEM.AddEvent(e)
}

// GetEvent get event by name.
func GetEvent(name string) (Event, bool) {
	return DefaultEM.GetEvent(name)
}

// HasEvent has event check.
func HasEvent(name string) bool {
	return DefaultEM.HasEvent(name)
}
