package integration

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_VersionOutput(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("Version")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, "1.40.1", out)
	}

	t.Log("Version --full")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version", "--full")
		require.NoError(t, err)

		expectedOSVersion := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		expectedVersionOut := fmt.Sprintf(`version: 1.40.1
format version: 10
os: %s
go: %s
build number: 
commit:`, expectedOSVersion, runtime.Version())

		require.Equal(t, expectedVersionOut, out)
	}
}
