package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestUnsetCmd_ClearsValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{APIBaseURL: "https://api.example.com"}))

	cmd, out := newTestCmd(t, NewUnsetCommand())
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL}))

	assert.Equal(t, "Cleared api_base_url\n", out.String())

	cfg, err := internalconfig.Load()
	require.NoError(t, err)
	assert.Equal(t, "", cfg.APIBaseURL)
}

func TestUnsetCmd_UnknownKeyErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newTestCmd(t, NewUnsetCommand())
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.RunE(cmd, []string{"nope"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestUnsetCmd_RequiresExactlyOneArg(t *testing.T) {
	cmd := NewUnsetCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())
}

func TestUnsetCmd_ExecuteRejectsUnknownKeyBeforeRunE(t *testing.T) {
	cmd := NewUnsetCommand()
	cmd.SetArgs([]string{"nope"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid argument "nope"`)
}
