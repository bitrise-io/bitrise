package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

func TestResolveTokenAndSource_NoToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tok, source := resolveTokenAndSource()
	assert.Empty(t, tok)
	assert.Equal(t, "none", source)
}

func TestResolveTokenAndSource_EnvTakesPrecedence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "file-token"}))
	t.Setenv(auth.EnvToken, "env-token")

	tok, source := resolveTokenAndSource()
	assert.Equal(t, "env-token", tok)
	assert.Equal(t, "env (BITRISE_TOKEN)", source)
}

func TestResolveTokenAndSource_FallsBackToAuthFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "file-token"}))

	tok, source := resolveTokenAndSource()
	assert.Equal(t, "file-token", tok)
	assert.Equal(t, "auth file", source)
}

func TestCurrentStatus_NoToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	s, err := currentStatus()
	require.NoError(t, err)
	assert.False(t, s.HasToken)
	assert.Empty(t, s.TokenType)
	assert.Equal(t, "none", s.Source)
}

func TestCurrentStatus_PastedToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "bitpat_x"}))

	s, err := currentStatus()
	require.NoError(t, err)
	assert.True(t, s.HasToken)
	assert.Equal(t, "PAT", s.TokenType)
	assert.Equal(t, "auth file", s.Source)
	assert.Empty(t, s.TokenExpiry)
}

func TestCurrentStatus_OAuthManagedShowsExpiry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	require.NoError(t, auth.Save(auth.Auth{
		Token: "bitpat_x", TokenExpiry: expiry,
		JWT: "jwt", JWTExpiry: expiry, RefreshToken: "refresh",
	}))

	s, err := currentStatus()
	require.NoError(t, err)
	assert.Equal(t, "oauth (auth file)", s.Source)
	assert.Equal(t, expiry.Format(time.RFC3339), s.TokenExpiry)
}

func TestCurrentStatus_EnvTokenSkipsOAuthDetails(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "bitpat_x", TokenExpiry: time.Now().Add(time.Hour),
		JWT: "jwt", JWTExpiry: time.Now().Add(time.Hour), RefreshToken: "refresh",
	}))
	t.Setenv(auth.EnvToken, "env-token")

	s, err := currentStatus()
	require.NoError(t, err)
	assert.Equal(t, "env (BITRISE_TOKEN)", s.Source)
	assert.Empty(t, s.TokenExpiry)
}
