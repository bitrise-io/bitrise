package integration

import (
	"os"
	"os/exec"
	"testing"

	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Update(t *testing.T) {

	t.Log("Update --version 1.10.0")
	{
		// save original binary
		if err := copyFile(binPath(), binPath()+"_original"); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath(), "update", "--version", "1.10.0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		require.NoError(t, err)

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, "1.10.0", updatedVer)

		// restore original binary
		if err := copyFile(binPath()+"_original", binPath()); err != nil {
			t.Fatal(err)
		}

		restoredVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.Equal(t, version.VERSION, restoredVer)
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
