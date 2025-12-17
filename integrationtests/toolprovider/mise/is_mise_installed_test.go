//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapSkipsInstallationWhenMiseAlreadyInstalled(t *testing.T) {
	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, false)
	require.NoError(t, err)

	// Should install
	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	// Exists and is executable test
	misePath := filepath.Join(miseInstallDir, "bin", "mise")
	info, err := os.Stat(misePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
	assert.NotEqual(t, 0, info.Mode().Perm()&0111, "mise binary should be executable")

	// Capture modification time, so we can verify it doesn't change on re-bootstrap
	originalModTime := info.ModTime()
	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	info, err = os.Stat(misePath)
	require.NoError(t, err)
	assert.Equal(t, originalModTime, info.ModTime(), "mise binary should not have been reinstalled")
}
