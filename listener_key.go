// Package event
package event

import (
	"fmt"
	"reflect"
	"unsafe"
	_ "unsafe" // go linkname
)

func getReflectRawPointer(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Func:
		return uintptr(getNativePointerLink(&v))
	case reflect.Pointer, reflect.Chan, reflect.Map, reflect.UnsafePointer:
		return v.Pointer()
	default:
		return 0
	}
}

// only validate in reflect.Pointer, reflect.Chan, reflect.Map, reflect.UnsafePointer, reflect.Func
//go:linkname getNativePointerLink  reflect.(*Value).pointer
func getNativePointerLink(*reflect.Value) unsafe.Pointer

type listenerCompareKey struct {
	strData string
	ptrData uintptr
}

func getListenCompareKey(src Listener) listenerCompareKey {
	ret := listenerCompareKey{
		ptrData: getReflectRawPointer(reflect.ValueOf(src)),
	}
	if ret.ptrData == 0 {
		//ret.strData = fmt.Sprintf("%v", src)
		ret.strData = fmt.Sprintf("%p", any(src))
	}
	return ret
}
