//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"os"
	"os/exec"
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// The oldest release that publishes a binary for all four supported platforms.
const updateTestVersion = "2.43.2"

func Test_Update(t *testing.T) {
	t.Run("update --version "+updateTestVersion, func(t *testing.T) {
		keepBinary(t)

		out, err := command.New(testhelpers.BinPath(), "update", "--version", updateTestVersion).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)
		require.Equal(t, updateTestVersion, updatedVer)
	})

	t.Run("update --version invalid", func(t *testing.T) {
		originalVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)

		require.Error(t, exec.Command(testhelpers.BinPath(), "update", "--version", "invalid").Run())

		updatedVer, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version")
		require.NoError(t, err)
		require.Equal(t, originalVer, updatedVer)
	})
}

// The binary under test is the one the update replaces, so the rest of the suite
// needs it put back, failure or not.
func keepBinary(t *testing.T) {
	original, err := os.ReadFile(testhelpers.BinPath())
	require.NoError(t, err)
	info, err := os.Stat(testhelpers.BinPath())
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.WriteFile(testhelpers.BinPath(), original, info.Mode().Perm()))
	})
}
