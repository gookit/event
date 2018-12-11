package event

// Listener interface
type Listener interface {
	Handle(e Event) error
}

// ListenerFunc func definition.
type ListenerFunc func(e Event) error

// Handle event. implements the Listener interface
func (fn ListenerFunc) Handle(e Event) error {
	return fn(e)
}
