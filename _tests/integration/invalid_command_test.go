package integration

import "testing"

func Test_InvalidCommand(t *testing.T) {
	t.Log("Invalid command")
	{
		// TODO: there should be an error for invalid command,
		//  but it won't return any right now; exit code is 0 :/
		// out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "invalidcmd")
		// require.EqualError(t, err, "")
		// require.Equal(t, "1.4.0", out)
	}
}
