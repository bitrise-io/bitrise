//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/cli"
	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Update(t *testing.T) {
	// update replaces the binary it runs from, so each case runs a copy: the other tests,
	// running in parallel, need the binary under test.
	t.Log("Update --version 2.31.0")
	{
		binPath := copyBinary(t)

		out, err := command.New(binPath, "update", "--version", "2.31.0").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, "version")
		require.NoError(t, err)
		require.Equal(t, "2.31.0", updatedVer)
	}
	t.Log("Update --version invalid")
	{
		binPath := copyBinary(t)

		originalVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, "version")
		require.NoError(t, err)
		err = exec.Command(binPath, "update", "--version", "invalid").Run()
		require.Error(t, err)
		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, "version")
		require.NoError(t, err)
		require.Equal(t, originalVer, updatedVer)
	}
}

func copyBinary(t *testing.T) string {
	binPath := filepath.Join(t.TempDir(), "bitrise")
	if err := cli.CopyFile(testhelpers.BinPath(), binPath, false); err != nil {
		t.Fatal(err)
	}
	return binPath
}
