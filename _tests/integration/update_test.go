package integration

import (
	"os/exec"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Update(t *testing.T) {
	t.Log("Update --version 1.10.0")
	{
		err := exec.Command(binPath(), "update", "--version", "1.10.0").Run()
		require.NoError(t, err)
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.Equal(t, "1.10.0", out)
	}
	t.Log("Update --version invalid")
	{
		originalVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		err = exec.Command(binPath(), "update", "--version", "invalid").Run()
		require.Error(t, err)
		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, originalVer, updatedVer)
	}
}
