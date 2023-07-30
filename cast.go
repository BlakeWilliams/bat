package bat

import (
	"fmt"
	"reflect"
)

func castInt64(given reflect.Value) reflect.Value {
	switch given.Type().Kind() {
	case reflect.Int:
		return reflect.ValueOf(given.Int())
	default:
		panic(fmt.Sprintf("castInt64 does not support %s", given.Type().Kind()))
	}
}
