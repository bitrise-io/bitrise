package auth

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
	assert.Equal(t, filepath.Join("/custom/xdg", "bitrise", "auth.yaml"), got)
}

func TestPath_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := Path()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(got, filepath.Join(".config", "bitrise", "auth.yaml")))
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
		p := filepath.Join(dir, "bitrise", "auth.yaml")
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

func TestSave_RejectsEmptyToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	err := Save(Auth{Token: ""})
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	bad := filepath.Join(dir, "bitrise", "auth.yaml")
	require.NoError(t, os.WriteFile(bad, []byte("this: is :: bad yaml\n: oops"), 0o600))
	_, err := Load()
	assert.Error(t, err)
}

func TestSaveLoad_OAuthFields_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// UTC + second precision avoids monotonic-clock / location drift through YAML.
	want := Auth{
		Token:              "bitpat_x",
		TokenExpiry:        time.Now().Add(time.Hour).UTC().Truncate(time.Second),
		JWT:                "header.payload.sig",
		JWTExpiry:          time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second),
		RefreshToken:       "refresh-1",
		RefreshTokenExpiry: time.Now().Add(720 * time.Hour).UTC().Truncate(time.Second),
	}
	require.NoError(t, Save(want))
	got, err := Load()
	require.NoError(t, err)

	assert.Equal(t, want.Token, got.Token)
	assert.Equal(t, want.JWT, got.JWT)
	assert.Equal(t, want.RefreshToken, got.RefreshToken)
	assert.True(t, got.TokenExpiry.Equal(want.TokenExpiry))
	assert.True(t, got.JWTExpiry.Equal(want.JWTExpiry))
	assert.True(t, got.RefreshTokenExpiry.Equal(want.RefreshTokenExpiry))
	assert.True(t, got.IsOAuthManaged())
}

func TestIsOAuthManaged(t *testing.T) {
	assert.False(t, (Auth{Token: "x"}).IsOAuthManaged())
	assert.True(t, (Auth{Token: "x", RefreshToken: "r"}).IsOAuthManaged())
}

func TestLoad_BackwardCompat_TokenOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	p := filepath.Join(dir, "bitrise", "auth.yaml")
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
