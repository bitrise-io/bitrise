package bitrise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupShims(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	exe, err := os.Executable()
	require.NoError(t, err)
	shim := filepath.Join(configs.GetBitriseShimsDirPath(), "envman")

	require.NoError(t, SetupShims(true))
	require.NoError(t, SetupShims(true))
	target, err := os.Readlink(shim)
	require.NoError(t, err)
	assert.Equal(t, exe, target)

	require.NoError(t, SetupShims(false))
	_, err = os.Lstat(shim)
	assert.ErrorIs(t, err, os.ErrNotExist)
	require.NoError(t, SetupShims(false), "removing a missing shim is not an error")
}

func TestRefreshShims(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	exe, err := os.Executable()
	require.NoError(t, err)
	shim := filepath.Join(configs.GetBitriseShimsDirPath(), "envman")

	require.NoError(t, RefreshShims())
	_, err = os.Lstat(shim)
	assert.ErrorIs(t, err, os.ErrNotExist, "no shim is created without the opt-in")

	require.NoError(t, os.MkdirAll(filepath.Dir(shim), 0o755))
	require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "removed-cli"), shim))
	require.NoError(t, RefreshShims())
	target, err := os.Readlink(shim)
	require.NoError(t, err)
	assert.Equal(t, exe, target, "a dangling shim points at the running CLI again")
}

func TestLookPathOutsideShims(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	shimsDir := configs.GetBitriseShimsDirPath()
	installedDir := t.TempDir()
	notExecutableDir := t.TempDir()
	for _, dir := range []string{shimsDir, installedDir} {
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "envman"), []byte("#!/bin/sh\n"), 0o755))
	}
	require.NoError(t, os.WriteFile(filepath.Join(notExecutableDir, "envman"), []byte(""), 0o644))

	t.Setenv("PATH", shimsDir+":"+notExecutableDir+":"+installedDir)
	pth, err := lookPathOutsideShims("envman")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(installedDir, "envman"), pth)

	// A bare "envman" from a "." entry would be looked up on PATH again when run.
	t.Chdir(installedDir)
	t.Setenv("PATH", shimsDir+":.")
	pth, err = lookPathOutsideShims("envman")
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(pth), pth)
	assert.Equal(t, "envman", filepath.Base(pth))

	t.Setenv("PATH", shimsDir)
	_, err = lookPathOutsideShims("envman")
	assert.EqualError(t, err, "envman not found in PATH")
}
