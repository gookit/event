package simpleevent

// DefaultEM default event manager
var DefaultEM = NewEventManager()

// On register a event and handler
func On(name string, handler HandlerFunc) {
	DefaultEM.On(name, handler)
}

// Has event check.
func Has(name string) bool {
	return DefaultEM.HasEvent(name)
}

// Fire fire handlers by name.
func Fire(name string, args ...interface{}) error {
	return DefaultEM.Fire(name, args)
}

// MustFire fire event by name. will panic on error
func MustFire(name string, args ...interface{}) {
	DefaultEM.MustFire(name, args)
}
