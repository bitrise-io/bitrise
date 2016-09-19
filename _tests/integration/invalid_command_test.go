package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_InvalidCommand(t *testing.T) {
	t.Log("Invalid command")
	{
		_, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "invalidcmd")
		require.EqualError(t, err, "exit status 1")
	}
}
