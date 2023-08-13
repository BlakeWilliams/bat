package bat

import (
	"fmt"
	"math"
	"reflect"
)

// These functions are somehat naive and assumes that the right-most type
// should be the cast target. A more comprehensive implementation
// would be very welcome.

func subtract(a any, b any) any {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if !aValue.CanConvert(bValue.Type()) {
		panic(fmt.Sprintf("can't convert type %s into %s", aValue.Type(), bValue.Type()))
	}

	switch reflect.ValueOf(b).Kind() {
	case reflect.Int64:
		return a.(int64) - b.(int64)
	case reflect.Int32:
		return a.(int32) - b.(int32)
	case reflect.Int16:
		return a.(int16) - b.(int16)
	case reflect.Int8:
		return a.(int8) - b.(int8)
	case reflect.Int:
		return a.(int) - b.(int)
	case reflect.Uint64:
		return a.(uint64) - b.(uint64)
	case reflect.Uint32:
		return a.(uint32) - b.(uint32)
	case reflect.Uint16:
		return a.(uint16) - b.(uint16)
	case reflect.Uint8:
		return a.(uint8) - b.(uint8)
	case reflect.Uint:
		return a.(uint) - b.(uint)
	case reflect.Float32:
		return a.(float32) - b.(float32)
	case reflect.Float64:
		return a.(float64) - b.(float64)
	case reflect.Complex64:
		return a.(complex64) - b.(complex64)
	case reflect.Complex128:
		return a.(complex128) - b.(complex128)
	default:
		panic(fmt.Sprintf("can't subtract %s from %s", aValue.Kind(), bValue.Kind()))
	}
}

func add(a any, b any, escapeFunc func(string) string) any {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if !aValue.CanConvert(bValue.Type()) {
		panic(fmt.Sprintf("can't convert type %s into %s", aValue.Type(), bValue.Type()))
	}

	if aValue.Kind() == reflect.String {
		left := aValue.String()
		right := bValue.String()

		if aValue.Type().Name() != "Safe" {
			left = escapeFunc(left)
		}

		if bValue.Type().Name() != "Safe" {
			right = escapeFunc(right)
		}

		return Safe(left + right)
	}

	switch reflect.ValueOf(b).Kind() {
	case reflect.Int64:
		return a.(int64) + b.(int64)
	case reflect.Int32:
		return a.(int32) + b.(int32)
	case reflect.Int16:
		return a.(int16) + b.(int16)
	case reflect.Int8:
		return a.(int8) + b.(int8)
	case reflect.Int:
		return a.(int) + b.(int)
	case reflect.Uint64:
		return a.(uint64) + b.(uint64)
	case reflect.Uint32:
		return a.(uint32) + b.(uint32)
	case reflect.Uint16:
		return a.(uint16) + b.(uint16)
	case reflect.Uint8:
		return a.(uint8) + b.(uint8)
	case reflect.Uint:
		return a.(uint) + b.(uint)
	case reflect.Float32:
		return a.(float32) + b.(float32)
	case reflect.Float64:
		return a.(float64) + b.(float64)
	case reflect.Complex64:
		return a.(complex64) + b.(complex64)
	case reflect.Complex128:
		return a.(complex128) + b.(complex128)
	default:
		panic(fmt.Sprintf("can't add %s from %s", aValue.Kind(), bValue.Kind()))
	}
}

func multiply(a any, b any) any {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if !aValue.CanConvert(bValue.Type()) {
		panic(fmt.Sprintf("can't convert type %s into %s", aValue.Type(), bValue.Type()))
	}

	switch reflect.ValueOf(b).Kind() {
	case reflect.Int64:
		return a.(int64) * b.(int64)
	case reflect.Int32:
		return a.(int32) * b.(int32)
	case reflect.Int16:
		return a.(int16) * b.(int16)
	case reflect.Int8:
		return a.(int8) * b.(int8)
	case reflect.Int:
		return a.(int) * b.(int)
	case reflect.Uint64:
		return a.(uint64) * b.(uint64)
	case reflect.Uint32:
		return a.(uint32) * b.(uint32)
	case reflect.Uint16:
		return a.(uint16) * b.(uint16)
	case reflect.Uint8:
		return a.(uint8) * b.(uint8)
	case reflect.Uint:
		return a.(uint) * b.(uint)
	case reflect.Float32:
		return a.(float32) * b.(float32)
	case reflect.Float64:
		return a.(float64) * b.(float64)
	case reflect.Complex64:
		return a.(complex64) * b.(complex64)
	case reflect.Complex128:
		return a.(complex128) * b.(complex128)
	default:
		panic(fmt.Sprintf("can't subtract %s from %s", aValue.Kind(), bValue.Kind()))
	}
}

func divide(a any, b any) any {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if !aValue.CanConvert(bValue.Type()) {
		panic(fmt.Sprintf("can't convert type %s into %s", aValue.Type(), bValue.Type()))
	}

	switch reflect.ValueOf(b).Kind() {
	case reflect.Int64:
		return a.(int64) / b.(int64)
	case reflect.Int32:
		return a.(int32) / b.(int32)
	case reflect.Int16:
		return a.(int16) / b.(int16)
	case reflect.Int8:
		return a.(int8) / b.(int8)
	case reflect.Int:
		return a.(int) / b.(int)
	case reflect.Uint64:
		return a.(uint64) / b.(uint64)
	case reflect.Uint32:
		return a.(uint32) / b.(uint32)
	case reflect.Uint16:
		return a.(uint16) / b.(uint16)
	case reflect.Uint8:
		return a.(uint8) / b.(uint8)
	case reflect.Uint:
		return a.(uint) / b.(uint)
	case reflect.Float32:
		return a.(float32) / b.(float32)
	case reflect.Float64:
		return a.(float64) / b.(float64)
	case reflect.Complex64:
		return a.(complex64) / b.(complex64)
	case reflect.Complex128:
		return a.(complex128) / b.(complex128)
	default:
		panic(fmt.Sprintf("can't subtract %s from %s", aValue.Kind(), bValue.Kind()))
	}
}

func modulo(a any, b any) any {
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if !aValue.CanConvert(bValue.Type()) {
		panic(fmt.Sprintf("can't convert type %s into %s", aValue.Type(), bValue.Type()))
	}

	switch reflect.ValueOf(b).Kind() {
	case reflect.Int64:
		return a.(int64) % b.(int64)
	case reflect.Int32:
		return a.(int32) % b.(int32)
	case reflect.Int16:
		return a.(int16) % b.(int16)
	case reflect.Int8:
		return a.(int8) % b.(int8)
	case reflect.Int:
		return a.(int) % b.(int)
	case reflect.Uint64:
		return a.(uint64) % b.(uint64)
	case reflect.Uint32:
		return a.(uint32) % b.(uint32)
	case reflect.Uint16:
		return a.(uint16) % b.(uint16)
	case reflect.Uint8:
		return a.(uint8) % b.(uint8)
	case reflect.Uint:
		return a.(uint) % b.(uint)
	case reflect.Float32:
		return math.Mod(a.(float64), b.(float64))
	case reflect.Float64:
		return math.Mod(a.(float64), b.(float64))
	default:
		panic(fmt.Sprintf("can't subtract %s from %s", aValue.Kind(), bValue.Kind()))
	}
}
