package stringutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxLastChars(t *testing.T) {
	require.Equal(t, "", MaxLastChars("", 10))
	require.Equal(t, "a", MaxLastChars("a", 1))
	require.Equal(t, "a", MaxLastChars("ba", 1))
	require.Equal(t, "ba", MaxLastChars("ba", 10))
	require.Equal(t, "a", MaxLastChars("cba", 1))
	require.Equal(t, "cba", MaxLastChars("cba", 10))

	require.Equal(t, "llo world!", MaxLastChars("hello world!", 10))
}
