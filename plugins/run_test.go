package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrip(t *testing.T) {
	str := "test case"
	require.Equal(t, "test case", strip(str))

	str = " test case"
	require.Equal(t, "test case", strip(str))

	str = "test case "
	require.Equal(t, "test case", strip(str))

	str = "   test case   "
	require.Equal(t, "test case", strip(str))

	str = ""
	require.Equal(t, "", strip(str))
}
