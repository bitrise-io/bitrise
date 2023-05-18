package integration

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_VersionOutput(t *testing.T) {
	t.Log("Version --full")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version", "--full")
		require.NoError(t, err)

		expectedOSVersion := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		expectedVersionOut := fmt.Sprintf(`version: %s
format version: 12
os: %s
go: %s
build number: 
commit: (devel)`, version.VERSION, expectedOSVersion, runtime.Version())

		require.Equal(t, expectedVersionOut, out)
	}
}
