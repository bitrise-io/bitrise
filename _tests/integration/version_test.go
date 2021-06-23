package integration

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

const (
	expectedCLIVersion = "1.47.2"
)

func Test_VersionOutput(t *testing.T) {
	t.Log("Version")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, expectedCLIVersion, out)
	}

	t.Log("Version --full")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version", "--full")
		require.NoError(t, err)

		expectedOSVersion := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		expectedVersionOut := fmt.Sprintf(`version: %s
format version: 11
os: %s
go: %s
build number: 
commit:`, expectedCLIVersion, expectedOSVersion, runtime.Version())

		require.Equal(t, expectedVersionOut, out)
	}
}
