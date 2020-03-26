package plugins

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/stretchr/testify/require"
)

func TestStrip(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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
