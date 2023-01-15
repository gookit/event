// Package event is lightweight event manager and dispatcher implements by Go.
package event

import (
	"reflect"
)

func getListenCompareKey(src Listener) reflect.Value {
	return reflect.ValueOf(src)
}
