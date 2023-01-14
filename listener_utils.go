package event

import (
	"fmt"
	"reflect"
	"unsafe"
)

func getReflectRawPointer[T any](src T) uintptr {
	//if src == nil {
	//	return 0
	//}
	v := reflect.ValueOf(src)
	switch v.Kind() {
	case reflect.Func:
		return getNativePointer(v)
	case reflect.Pointer, reflect.Chan, reflect.Map, reflect.UnsafePointer:
		return v.Pointer()
	default:
		return 0
	}
}

// only validate in reflect.Pointer, reflect.Chan, reflect.Map, reflect.UnsafePointer, reflect.Func
func getNativePointer[T any](src T) uintptr {
	vv := (*emptyInterface)(unsafe.Pointer(&src))
	return uintptr(vv.word)
}

// source code from reflect.ValueOf()
type emptyInterface struct {
	typ  *struct{} //*rtype
	word unsafe.Pointer
}

type anyRawValue struct {
	strData string
	ptrData uintptr
}

func getAnyRawValue[T any](src T) anyRawValue {
	ret := anyRawValue{
		ptrData: getReflectRawPointer(src),
	}
	if ret.ptrData == 0 {
		//ret.strData = fmt.Sprintf("%v", src)
		ret.strData = fmt.Sprintf("%p", any(src))
	}
	return ret
}

func getListenCompareKey(src Listener) anyRawValue {
	return getAnyRawValue(src)
}
