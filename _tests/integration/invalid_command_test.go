package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_InvalidCommand(t *testing.T) {
	t.Log("Invalid command")
	{
		_, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "invalidcmd")
		require.EqualError(t, err, "exit status 1")
	}
}
