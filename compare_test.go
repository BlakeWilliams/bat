package bat

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCompare(t *testing.T) {
	testCases := map[string]struct {
		left     any
		right    any
		expected bool
	}{
		"nil is equal to nil": {
			left:     nil,
			right:    nil,
			expected: true,
		},
		"nil is equal to nil pointer": {
			left:     nil,
			right:    (*int)(nil),
			expected: true,
		},
		"nil is not equal to time.Time{}": {
			left:     nil,
			right:    time.Time{},
			expected: false,
		},
		"equal strings return true": {
			left:     "foo",
			right:    "foo",
			expected: true,
		},
		"unequal strings return true": {
			left:     "foo",
			right:    "bar",
			expected: false,
		},
		"3 is not equal to 4": {
			left:     3,
			right:    4,
			expected: false,
		},
		"true is true": {
			left:     true,
			right:    true,
			expected: true,
		},
		"false is false": {
			left:     false,
			right:    false,
			expected: true,
		},
		"false is not true": {
			left:     true,
			right:    false,
			expected: false,
		},
	}
	for name, tC := range testCases {
		t.Run(name, func(t *testing.T) {
			result := compare(reflect.ValueOf(tC.left), reflect.ValueOf(tC.right))
			require.Equal(t, tC.expected, result)

			result = compare(reflect.ValueOf(tC.right), reflect.ValueOf(tC.left))
			require.Equal(t, tC.expected, result)
		})
	}
}
