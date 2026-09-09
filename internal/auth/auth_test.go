package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath_HonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/custom/xdg", "bitrise", "cli", "auth.yaml"), got)
}

func TestPath_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := Path()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, filepath.Join(".config", "bitrise", "cli", "auth.yaml")))
}

func TestSaveLoadClear_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Auth{}, got)

	want := Auth{Token: "secret-pat-123"}
	require.NoError(t, Save(want))
	got, err = Load()
	require.NoError(t, err)
	assert.Equal(t, want, got)

	if runtime.GOOS != "windows" {
		p := filepath.Join(dir, "bitrise", "cli", "auth.yaml")
		info, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		dirInfo, err := os.Stat(filepath.Dir(p))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	}

	require.NoError(t, Clear())
	got, err = Load()
	require.NoError(t, err)
	assert.Equal(t, Auth{}, got)

	// Clear is idempotent.
	require.NoError(t, Clear())
}

func TestLoad_FallsBackToPredecessorAuth(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	predecessorDir := filepath.Join(dir, "bitrise")
	require.NoError(t, os.MkdirAll(predecessorDir, 0o700))
	predecessorFile := filepath.Join(predecessorDir, "auth.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte("token: bitpat_predecessor\n"), 0o600))

	fallbackAnnounceOnce = sync.Once{}
	var announced strings.Builder
	origWriter := fallbackWriter
	fallbackWriter = &announced
	t.Cleanup(func() { fallbackWriter = origWriter })

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_predecessor", got.Token)

	_, err = os.ReadFile(predecessorFile)
	require.NoError(t, err, "the predecessor auth file must survive Load — it's read-only")
	_, statErr := os.Stat(filepath.Join(dir, "bitrise", "cli", "auth.yaml"))
	assert.True(t, os.IsNotExist(statErr), "Load must not write the new auth file as a side effect")

	assert.Contains(t, announced.String(), predecessorFile)
}

func TestLoad_PrefersNewAuthWhenPresent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bitrise", "auth.yaml"), []byte("token: bitpat_predecessor\n"), 0o600))
	require.NoError(t, Save(Auth{Token: "bitpat_new"}))

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_new", got.Token, "the new file must win once it exists, regardless of the predecessor file")
}

func TestActivePath_PredecessorWhileFallbackLive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	predecessorFile := filepath.Join(dir, "bitrise", "auth.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte("token: bitpat_predecessor\n"), 0o600))

	got, err := ActivePath()
	require.NoError(t, err)
	assert.Equal(t, predecessorFile, got)
}

func TestActivePath_NewPathOnceWritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bitrise", "auth.yaml"), []byte("token: bitpat_predecessor\n"), 0o600))
	require.NoError(t, Save(Auth{Token: "bitpat_new"}))

	got, err := ActivePath()
	require.NoError(t, err)
	newPath, err := Path()
	require.NoError(t, err)
	assert.Equal(t, newPath, got, "once the new file has been written, ActivePath must name it even though the predecessor file still exists")
}

func TestActivePath_NewPathWhenNeitherExists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := ActivePath()
	require.NoError(t, err)
	newPath, err := Path()
	require.NoError(t, err)
	assert.Equal(t, newPath, got)
}

func TestClear_RemovesBothCurrentAndPredecessorFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	predecessorFile := filepath.Join(dir, "bitrise", "auth.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte("token: bitpat_predecessor\n"), 0o600))
	require.NoError(t, Save(Auth{Token: "bitpat_new"}))

	require.NoError(t, Clear())

	_, err := os.Stat(predecessorFile)
	assert.True(t, os.IsNotExist(err), "Clear must remove the predecessor file too, or logout followed by status would fall back and report still logged in")
	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Auth{}, got)
}

func TestSave_RejectsEmptyToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	err := Save(Auth{Token: ""})
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise", "cli"), 0o700))
	bad := filepath.Join(dir, "bitrise", "cli", "auth.yaml")
	require.NoError(t, os.WriteFile(bad, []byte("this: is :: bad yaml\n: oops"), 0o600))
	_, err := Load()
	assert.Error(t, err)
}

func TestSaveLoad_OAuthFields_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// UTC + second precision avoids monotonic-clock / location drift through YAML.
	want := Auth{
		Token:        "bitpat_x",
		TokenExpiry:  time.Now().Add(time.Hour).UTC().Truncate(time.Second),
		JWT:          "header.payload.sig",
		JWTExpiry:    time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second),
		RefreshToken: "refresh-1",
	}
	require.NoError(t, Save(want))
	got, err := Load()
	require.NoError(t, err)

	assert.Equal(t, want.Token, got.Token)
	assert.Equal(t, want.JWT, got.JWT)
	assert.Equal(t, want.RefreshToken, got.RefreshToken)
	assert.True(t, got.TokenExpiry.Equal(want.TokenExpiry))
	assert.True(t, got.JWTExpiry.Equal(want.JWTExpiry))
	assert.True(t, got.IsOAuthManaged())
}

func TestIsOAuthManaged(t *testing.T) {
	assert.False(t, (Auth{Token: "x"}).IsOAuthManaged())
	assert.True(t, (Auth{Token: "x", RefreshToken: "r"}).IsOAuthManaged())
}

func TestLoad_BackwardCompat_TokenOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	p := filepath.Join(dir, "bitrise", "cli", "auth.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
	// An auth.yaml written before OAuth support: only `token`.
	require.NoError(t, os.WriteFile(p, []byte("token: bitpat_old\n"), 0o600))

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_old", got.Token)
	assert.False(t, got.IsOAuthManaged())
}

// TestSave_OverwritesExisting verifies Save survives an existing file (atomic replace).
func TestSave_OverwritesExisting(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, Save(Auth{Token: "first"}))
	require.NoError(t, Save(Auth{Token: "second"}))
	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "second", got.Token)
}
