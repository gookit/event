package event

// DefaultEM default event manager
var DefaultEM = NewManager("default")

/*************************************************************
 * Listener
 *************************************************************/

// On register a event and listener
func On(name string, listener Listener, priority int) {
	DefaultEM.On(name, listener, priority)
}

// Fire fire listeners by name.
func Fire(name string, args ...interface{}) error {
	return DefaultEM.Fire(name, args)
}

// FireEvent fire listeners by Event instance.
func FireEvent(e Event) error {
	return DefaultEM.FireEvent(e)
}

// MustFire fire event by name. will panic on error
func MustFire(name string, args ...interface{}) {
	DefaultEM.MustFire(name, args)
}

// HasListeners has listeners for the event name.
func HasListeners(name string) bool {
	return DefaultEM.HasListeners(name)
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
