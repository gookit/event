package simpleevent

import (
	"reflect"
	"runtime"
)

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
func Fire(name string, args ...any) error {
	return DefaultEM.Fire(name, args)
}

// MustFire fire event by name. will panic on error
func MustFire(name string, args ...any) {
	DefaultEM.MustFire(name, args)
}

func funcName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
