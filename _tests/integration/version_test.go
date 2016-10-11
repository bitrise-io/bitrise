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
		require.Equal(t, "1.4.1", out)
	}

	t.Log("Version --full")
	{
		out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version", "--full")
		require.NoError(t, err)
		require.Equal(t, `version: 1.4.1
format version: 1.3.0
build number:`+` `+`
commit:`, out)
	}
}
