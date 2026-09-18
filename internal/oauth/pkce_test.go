package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPKCE(t *testing.T) {
	verifier, challenge, err := newPKCE()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)
	assert.NotEmpty(t, challenge)

	_, err = base64.RawURLEncoding.DecodeString(verifier)
	require.NoError(t, err, "verifier should be raw base64url")

	sum := sha256.Sum256([]byte(verifier))
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), challenge, "challenge should be S256(verifier)")

	verifier2, _, err := newPKCE()
	require.NoError(t, err)
	assert.NotEqual(t, verifier, verifier2, "verifier should be random")
}

func TestNewState(t *testing.T) {
	s1, err := newState()
	require.NoError(t, err)
	assert.NotEmpty(t, s1)

	s2, err := newState()
	require.NoError(t, err)
	assert.NotEqual(t, s1, s2, "state should be random")
}
