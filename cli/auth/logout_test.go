package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

func TestRunLogout_ClearsAuthFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "bitpat_x"}))

	p, err := auth.Path()
	require.NoError(t, err)
	_, statErr := os.Stat(p)
	require.NoError(t, statErr, "auth.yaml should exist before logout")

	require.NoError(t, runLogout())

	_, statErr = os.Stat(p)
	assert.True(t, os.IsNotExist(statErr), "auth.yaml should be removed after logout")
}

func TestRunLogout_NoFileIsNotAnError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	require.NoError(t, runLogout())
}
