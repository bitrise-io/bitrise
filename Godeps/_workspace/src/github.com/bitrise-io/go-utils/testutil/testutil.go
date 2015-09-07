package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// EqualAndNoError ...
func EqualAndNoError(t *testing.T, expected, actual interface{}, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
	require.Equal(t, expected, actual, msgAndArgs...)
}
