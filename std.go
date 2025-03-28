package event

// std default event manager
var std = NewManager("default")

// Std get default event manager
func Std() *Manager {
	return std
}

// Config set default event manager options
func Config(fn ...OptionFn) {
	std.WithOptions(fn...)
}

/*************************************************************
 * Listener
 *************************************************************/

// On register a listener to the event. alias of Listen()
func On(name string, listener Listener, priority ...int) {
	std.On(name, listener, priority...)
}

// On register a listener to the event. trigger once
func Once(name string, listener Listener, priority ...int) {
	std.Once(name, listener, priority...)
}

// Listen register a listener to the event
func Listen(name string, listener Listener, priority ...int) {
	std.Listen(name, listener, priority...)
}

// Subscribe register a listener to the event
func Subscribe(sbr Subscriber) {
	std.Subscribe(sbr)
}

// AddSubscriber register a listener to the event
func AddSubscriber(sbr Subscriber) {
	std.AddSubscriber(sbr)
}

// AsyncFire simple async fire event by 'go' keywords
func AsyncFire(e Event) {
	std.AsyncFire(e)
}

// Async fire event by channel
func Async(name string, params M) {
	std.Async(name, params)
}

// FireAsync fire event by channel
func FireAsync(e Event) {
	std.FireAsync(e)
}

// CloseWait close chan and wait for all async events done.
func CloseWait() error {
	return std.CloseWait()
}

// Trigger alias of Fire
func Trigger(name string, params M) (error, Event) {
	return std.Fire(name, params)
}

// Fire listeners by name.
func Fire(name string, params M) (error, Event) {
	return std.Fire(name, params)
}

// FireEvent fire listeners by Event instance.
func FireEvent(e Event) error {
	return std.FireEvent(e)
}

// TriggerEvent alias of FireEvent
func TriggerEvent(e Event) error {
	return std.FireEvent(e)
}

// MustFire fire event by name. will panic on error
func MustFire(name string, params M) Event {
	return std.MustFire(name, params)
}

// MustTrigger alias of MustFire
func MustTrigger(name string, params M) Event {
	return std.MustFire(name, params)
}

// FireBatch fire multi event at once.
func FireBatch(es ...any) []error {
	return std.FireBatch(es...)
}

// HasListeners has listeners for the event name.
func HasListeners(name string) bool {
	return std.HasListeners(name)
}

// Reset the default event manager
func Reset() {
	std.Clear()
}

/*************************************************************
 * Event
 *************************************************************/

// AddEvent add a pre-defined event.
func AddEvent(e Event) {
	std.AddEvent(e)
}

// AddEventFc add a pre-defined event factory func to manager.
func AddEventFc(name string, fc FactoryFunc) {
	std.AddEventFc(name, fc)
}

// GetEvent get event by name.
func GetEvent(name string) (Event, bool) {
	return std.GetEvent(name)
}

// HasEvent has event check.
func HasEvent(name string) bool {
	return std.HasEvent(name)
}
