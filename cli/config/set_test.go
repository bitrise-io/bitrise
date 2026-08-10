package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestSetCmd_SavesValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newTestCmd(t, NewSetCommand())
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL, "https://api.example.com"}))

	assert.Equal(t, "Saved api_base_url\n", errOut.String())

	cfg, err := internalconfig.Load()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", cfg.APIBaseURL)
}

func TestSetCmd_InvalidURLErrorsWithoutSaving(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newTestCmd(t, NewSetCommand())
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL, "not-a-url"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be an absolute URL")

	cfg, loadErr := internalconfig.Load()
	require.NoError(t, loadErr)
	assert.Equal(t, "", cfg.APIBaseURL)
}

func TestSetCmd_UnknownKeyErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newTestCmd(t, NewSetCommand())
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.RunE(cmd, []string{"nope", "value"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestSetCmd_PreservesOtherKeys(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{WebBaseURL: "https://app.example.com"}))

	cmd, _ := newTestCmd(t, NewSetCommand())
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL, "https://api.example.com"}))

	cfg, err := internalconfig.Load()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", cfg.APIBaseURL)
	assert.Equal(t, "https://app.example.com", cfg.WebBaseURL)
}

func TestSetCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := NewSetCommand()
	cmd.SetArgs([]string{internalconfig.KeyAPIBaseURL})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())
}
