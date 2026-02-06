//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"os/exec"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli"
	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Update(t *testing.T) {

	t.Log("Update --version 2.31.0")
	{
		// save original binary
		if err := cli.CopyFile(testhelpers.BinPath(), testhelpers.BinPath()+"_original", false); err != nil {
			t.Fatal(err)
		}

		out, err := command.New(testhelpers.BinPath(), "update", "--version", "2.31.0").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)
		require.Equal(t, "2.31.0", updatedVer)

		// restore original binary
		if err := cli.CopyFile(testhelpers.BinPath()+"_original", testhelpers.BinPath(), true); err != nil {
			t.Fatal(err)
		}

		restoredVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.Equal(t, version.VERSION, restoredVer)
	}
	t.Log("Update --version invalid")
	{
		originalVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)
		err = exec.Command(testhelpers.BinPath(), "update", "--version", "invalid").Run()
		require.Error(t, err)
		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)
		require.Equal(t, originalVer, updatedVer)
	}
}

func Test_Update_PluginsPersist(t *testing.T) {
	t.Log("Verify plugins persist after update")
	{
		// initial plugin list
		preUpdateOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "plugin", "list")
		require.NoError(t, err)
		require.NotContains(t, preUpdateOut, "No installed plugin found", "Expected default plugins to be installed")

		if err := cli.CopyFile(testhelpers.BinPath(), testhelpers.BinPath()+"_original", false); err != nil {
			t.Fatal(err)
		}

		out, err := command.New(testhelpers.BinPath(), "update").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// verify after update
		postUpdateOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "plugin", "list")
		require.NoError(t, err)
		require.NotContains(t, postUpdateOut, "No installed plugin found", "Plugins should persist after update")

		require.Contains(t, postUpdateOut, "init", "Default 'init' plugin should be present")
		require.Contains(t, postUpdateOut, "step", "Default 'step' plugin should be present")
		require.Contains(t, postUpdateOut, "workflow-editor", "Default 'workflow-editor' plugin should be present")

		require.Contains(t, out, "Checking Bitrise Plugins", "Update should validate plugins during setup")

		// restore original
		if err := cli.CopyFile(testhelpers.BinPath()+"_original", testhelpers.BinPath(), true); err != nil {
			t.Fatal(err)
		}
	}
}
