package testutil

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-utils/builtinutil"
	"github.com/stretchr/testify/require"
)

// EqualAndNoError ...
func EqualAndNoError(t *testing.T, expected, actual interface{}, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
	require.Equal(t, expected, actual, msgAndArgs...)
}

func equalSlicesWithoutOrder(t *testing.T, expected, actual []interface{}, msgAndArgs ...interface{}) {
	if !builtinutil.DeepEqualSlices(expected, actual) {
		require.FailNow(t, fmt.Sprintf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual), msgAndArgs...)
	}
}

// EqualSlicesWithoutOrder - regardless the order, but same items
func EqualSlicesWithoutOrder(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	castedExpected, err := builtinutil.CastInterfaceToInterfaceSlice(expected)
	if err != nil {
		require.FailNow(t, fmt.Sprintf("'expected' is not a slice: %#v", expected), msgAndArgs...)
	}

	castedActual, err := builtinutil.CastInterfaceToInterfaceSlice(actual)
	if err != nil {
		require.FailNow(t, fmt.Sprintf("'actual' is not a slice: %#v", actual), msgAndArgs...)
	}

	equalSlicesWithoutOrder(t, castedExpected, castedActual, msgAndArgs...)
}
