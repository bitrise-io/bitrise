package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_VersionOutput(t *testing.T) {
	t.Log("Version")
	{
		out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, "1.4.3", out)
	}

	t.Log("Version --full")
	{
		out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version", "--full")
		require.NoError(t, err)
		require.Equal(t, `version: 1.4.3
format version: 1.3.1
os: darwin (amd64)
go: go1.7.3
build number: 
commit:`, out)
	}
}
