package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestSetCmd_SavesValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, out := newTestCmd(t, NewSetCommand())
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL, "https://api.example.com"}))

	assert.Equal(t, "Saved api_base_url\n", out.String())

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

// TestSetCmd_CarriesPredecessorConfigForward covers the reason the legacy
// fallback lives in config.Load rather than only in config.Resolve: set is
// read-modify-write, so the first set after an upgrade has to read the
// predecessor CLI's file in full and write all of it forward. Reading only
// the (absent) new file would persist the one key just set and silently drop
// the rest, and the new file existing would then stop the fallback firing.
func TestSetCmd_CarriesPredecessorConfigForward(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	predecessorDir, err := internalconfig.PredecessorDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(predecessorDir, 0o700))
	legacyPath := filepath.Join(predecessorDir, "config.yaml")
	// app_slug/default_workspace_slug are the predecessor's pre-rename
	// spellings, so this also covers the key aliases.
	require.NoError(t, os.WriteFile(legacyPath, []byte(
		"app_slug: app-123\ndefault_workspace_slug: ws-456\ntheme: dark\nweb_base_url: https://app.example.com\n",
	), 0o600))

	cmd, _ := newTestCmd(t, NewSetCommand())
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.RunE(cmd, []string{internalconfig.KeyAPIBaseURL, "https://api.example.com"}))

	newPath, err := internalconfig.Path()
	require.NoError(t, err)
	require.FileExists(t, newPath)

	cfg, err := internalconfig.Load()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", cfg.APIBaseURL)
	assert.Equal(t, "app-123", cfg.AppID)
	assert.Equal(t, "ws-456", cfg.DefaultWorkspaceID)
	assert.Equal(t, "dark", cfg.Theme)
	assert.Equal(t, "https://app.example.com", cfg.WebBaseURL)

	// The predecessor CLI keeps running after the upgrade, so its file must
	// survive untouched.
	assert.FileExists(t, legacyPath)
	active, err := internalconfig.ActivePath()
	require.NoError(t, err)
	assert.Equal(t, newPath, active)
}

func TestSetCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := NewSetCommand()
	cmd.SetArgs([]string{internalconfig.KeyAPIBaseURL})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())
}

func TestSetCmd_ExecuteRejectsUnknownKeyBeforeRunE(t *testing.T) {
	cmd := NewSetCommand()
	cmd.SetArgs([]string{"nope", "value"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid argument "nope"`)
}

func TestSetCmd_ExecuteAllowsArbitraryValueArg(t *testing.T) {
	// The VALUE positional must NOT be checked against ValidArgs (only KEY
	// is) — cobra.OnlyValidArgs alone would wrongly reject any value that
	// isn't itself a recognized key.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewSetCommand()
	cmd.SetArgs([]string{internalconfig.KeyAPIBaseURL, "https://api.example.com"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
}
