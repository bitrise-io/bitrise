package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath_HonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/xdg", "bitrise", "cli", "config.yml"), got)
}

func TestPath_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := Path()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, filepath.Join(".config", "bitrise", "cli", "config.yml")))
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := Config{
		SetupVersion:           "2.1.0",
		LastCLIUpdateCheck:     time.Now().UTC().Truncate(time.Second),
		LastPluginUpdateChecks: map[string]time.Time{"init": time.Now().UTC().Truncate(time.Second)},
	}
	require.NoError(t, Save(want))

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, want.SetupVersion, got.SetupVersion)
	assert.True(t, got.LastCLIUpdateCheck.Equal(want.LastCLIUpdateCheck))
	assert.Len(t, got.LastPluginUpdateChecks, 1)
	assert.True(t, got.LastPluginUpdateChecks["init"].Equal(want.LastPluginUpdateChecks["init"]))

	if runtime.GOOS != "windows" {
		p := filepath.Join(dir, "bitrise", "cli", "config.yml")
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestLoad_MissingFileIsZeroValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Config{}, got)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise", "cli"), 0o700))
	bad := filepath.Join(dir, "bitrise", "cli", "config.yml")
	require.NoError(t, os.WriteFile(bad, []byte("this: is :: not yaml"), 0o600))
	_, err := Load()
	assert.Error(t, err)
}

func TestLoadDir_FindsAncestorFile(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	require.NoError(t, os.MkdirAll(deep, 0o755)) //nolint:gosec // test-only tempdir, perms don't matter

	cfgPath := filepath.Join(root, "a", DirFileName)
	require.NoError(t, os.WriteFile(cfgPath, []byte("setup_version: 1.2.3\n"), 0o644)) //nolint:gosec // test-only tempfile

	got, found, err := loadDirFrom(deep)
	require.NoError(t, err)
	assert.Equal(t, cfgPath, found)
	assert.Equal(t, "1.2.3", got.SetupVersion)
}

func TestLoadDir_NoFileReturnsZero(t *testing.T) {
	got, found, err := loadDirFrom(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, found)
	assert.Equal(t, Config{}, got)
}
