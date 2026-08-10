package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestGetCmd_PrintsStoredValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{APIBaseURL: "https://api.example.com"}))

	cmd, out := newTestCmd(t, NewGetCommand())
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL}))

	assert.Equal(t, "https://api.example.com\n", out.String())
}

func TestGetCmd_UnsetKeyPrintsEmptyLine(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, out := newTestCmd(t, NewGetCommand())
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyWebBaseURL}))

	assert.Equal(t, "\n", out.String())
}

func TestGetCmd_UnknownKeyErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newTestCmd(t, NewGetCommand())
	err := cmd.RunE(cmd, []string{"nope"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestGetCmd_RequiresExactlyOneArg(t *testing.T) {
	cmd := NewGetCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())

	cmd = NewGetCommand()
	cmd.SetArgs([]string{"a", "b"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())
}

func TestGetCmd_ExecuteRejectsUnknownKeyBeforeRunE(t *testing.T) {
	cmd := NewGetCommand()
	cmd.SetArgs([]string{"nope"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid argument "nope"`)
}
