package mapsort

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSort_String(t *testing.T) {
	m := map[string]string{"foo": "fooval", "bar": "barval"}

	sorted := Sort(reflect.ValueOf(m))

	resultKeys := make([]string, len(sorted.Keys))
	for i, key := range sorted.Keys {
		resultKeys[i] = key.Interface().(string)
	}

	resultValues := make([]string, len(sorted.Keys))
	for i, key := range sorted.Keys {
		resultValues[i] = key.Interface().(string)
	}

	require.Len(t, sorted.Keys, 2)
	require.Len(t, sorted.Values, 2)

	require.Equal(t, "bar", sorted.Keys[0].Interface())
	require.Equal(t, "foo", sorted.Keys[1].Interface())

	require.Equal(t, "barval", sorted.Values[0].Interface())
	require.Equal(t, "fooval", sorted.Values[1].Interface())
}
