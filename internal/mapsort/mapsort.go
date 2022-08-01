package mapsort

import (
	"reflect"
	"sort"
)

type Map struct {
	Keys   []reflect.Value
	Values []reflect.Value
}

func Sort(v reflect.Value) Map {
	len := v.Len()

	m := Map{
		Keys:   make([]reflect.Value, 0, len),
		Values: make([]reflect.Value, 0, len),
	}

	keyType := reflect.TypeOf(v.Interface()).Key()
	keys := v.MapKeys()

	if keyType.Comparable() {
		switch keyType.String() {
		case "string":
			sort.SliceStable(keys, func(a int, b int) bool {
				return keys[a].Interface().(string) < keys[b].Interface().(string)
			})
		}
	}

	for _, key := range keys {
		m.Keys = append(m.Keys, key)
		m.Values = append(m.Values, v.MapIndex(key))
	}

	return m
}
