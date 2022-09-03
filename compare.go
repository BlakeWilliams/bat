package bat

import (
	"reflect"
)

func compare(left reflect.Value, right reflect.Value) bool {
	if isNil(left) && isNil(right) {
		return true
	}

	if left.IsValid() && right.IsValid() {
		return left.Interface() == right.Interface()
	}

	return false
}

func isNil(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}

	return false
}

func isTruthy(v reflect.Value) bool {
	if isNil(v) {
		return false
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Invalid:
		return false
	default:
		return true
	}
}
