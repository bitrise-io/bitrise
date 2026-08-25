package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestListCmd_Empty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, out := newTestCmd(t, NewListCommand())
	require.NoError(t, cmd.RunE(cmd, nil))

	got := out.String()
	assert.Contains(t, got, "Path: ")
	assert.Contains(t, got, "api_base_url: (unset)")
	assert.Contains(t, got, "web_base_url: (unset)")
	assert.Contains(t, got, "rde_api_base_url: (unset)")
	assert.Contains(t, got, "app_id: (unset)")
	assert.Contains(t, got, "default_workspace_id: (unset)")
	assert.Contains(t, got, "output: (unset)")
	assert.Contains(t, got, "theme: (unset)")
}

func TestListCmd_ShowsSavedValues(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{APIBaseURL: "https://api.example.com"}))

	cmd, out := newTestCmd(t, NewListCommand())
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, out.String(), "api_base_url: https://api.example.com")
	assert.Contains(t, out.String(), "web_base_url: (unset)")
	assert.Contains(t, out.String(), "rde_api_base_url: (unset)")
	assert.Contains(t, out.String(), "app_id: (unset)")
	assert.Contains(t, out.String(), "default_workspace_id: (unset)")
	assert.Contains(t, out.String(), "output: (unset)")
	assert.Contains(t, out.String(), "theme: (unset)")
}

func TestListCmd_JSONFormat(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{APIBaseURL: "https://api.example.com"}))

	cmd, outBuf := newTestCmd(t, NewListCommand())
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))

	out := outBuf.String()
	assert.Contains(t, out, `"api_base_url": "https://api.example.com"`)
	assert.Contains(t, out, `"web_base_url": ""`, "unset keys must still be enumerated, not omitted")
	assert.Contains(t, out, `"rde_api_base_url": ""`)
	assert.Contains(t, out, `"app_id": ""`)
	assert.Contains(t, out, `"default_workspace_id": ""`)
	assert.Contains(t, out, `"output": ""`)
	assert.Contains(t, out, `"theme": ""`)
}

func TestListCmd_YMLFormat(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, internalconfig.Save(internalconfig.Config{APIBaseURL: "https://api.example.com"}))

	cmd, outBuf := newTestCmd(t, NewListCommand())
	require.NoError(t, cmd.Flags().Set("format", "yml"))
	require.NoError(t, cmd.RunE(cmd, nil))

	out := outBuf.String()
	assert.Contains(t, out, "api_base_url: https://api.example.com", "yaml.v2 ignores json tags; this key name regresses to \"apibaseurl\" without a yaml tag")
	assert.Contains(t, out, `web_base_url: ""`)
	assert.Contains(t, out, `rde_api_base_url: ""`)
	assert.Contains(t, out, `app_id: ""`)
	assert.Contains(t, out, `default_workspace_id: ""`)
	assert.Contains(t, out, `output: ""`)
	assert.Contains(t, out, `theme: ""`)
}

func TestListCmd_RejectsPositionalArgs(t *testing.T) {
	cmd := NewListCommand()
	cmd.SetArgs([]string{"unexpected"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "positional args should be rejected before the command runs")
}
