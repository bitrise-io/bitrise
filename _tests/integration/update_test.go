//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"os/exec"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Update(t *testing.T) {

	t.Log("Update --version 2.31.0")
	{
		// save original binary
		if err := cli.CopyFile(binPath(), binPath()+"_original", false); err != nil {
			t.Fatal(err)
		}

		out, err := command.New(binPath(), "update", "--version", "2.31.0").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "version")
		require.NoError(t, err)
		require.Equal(t, "2.31.0", updatedVer)

		// restore original binary
		if err := cli.CopyFile(binPath()+"_original", binPath(), true); err != nil {
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
