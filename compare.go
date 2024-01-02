package bat

import (
	"fmt"
	"reflect"
)

func compare(left reflect.Value, right reflect.Value) bool {
	if isNil(left) && isNil(right) {
		return true
	}

	if left.IsValid() && right.IsValid() {
		if left.Type() != right.Type() && right.Type().ConvertibleTo(left.Type()) {
			return left.Interface() == right.Convert(left.Type()).Interface()
		}
		return left.Interface() == right.Interface()
	}

	return false
}

func lessThan(leftValue any, rightValue any) (bool, error) {
	left := reflect.ValueOf(leftValue)
	right := reflect.ValueOf(rightValue)

	lKind := left.Kind()
	rKind := right.Kind()

	if lKind == rKind {
		switch lKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return left.Int() < right.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return left.Uint() < right.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return left.Float() < right.Float(), nil
		default:
			return false, fmt.Errorf("can't compare type %s", lKind)
		}
	}

	lCore := genericType(left)
	rCore := genericType(right)

	switch {
	case lCore == coreInt && rCore == coreUint:
		return uint64(left.Int()) < right.Uint(), nil
	case lCore == coreUint && rCore == coreInt:
		return left.Uint() < uint64(right.Int()), nil
	case lCore == coreFloat && rCore == coreInt:
		return left.Float() < float64(right.Int()), nil
	case lCore == coreInt && rCore == coreFloat:
		return float64(left.Int()) < right.Float(), nil
	case lCore == coreFloat && rCore == coreUint:
		return left.Float() < float64(right.Uint()), nil
	case lCore == coreUint && rCore == coreFloat:
		return float64(left.Uint()) < right.Float(), nil
	}

	return false, fmt.Errorf("can't compare type %s and %s", lKind, rKind)
}

func greaterThan(left any, right any) (bool, error) {
	return lessThan(right, left)
}

type coreType int

const (
	coreInvalid coreType = iota
	coreInt
	coreFloat
	coreUint
)

func genericType(v reflect.Value) coreType {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return coreInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return coreUint
	case reflect.Float32, reflect.Float64:
		return coreFloat
	default:
		return coreInvalid
	}
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
