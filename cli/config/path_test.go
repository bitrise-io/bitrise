package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathCmd_PrintsGlobalConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd, out := newTestCmd(t, NewPathCommand())
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, filepath.Join(dir, "bitrise", "cli", "config.yml")+"\n", out.String())
}

// TestPathCmd_PrintsPredecessorPathWhileFallbackLive covers the trap `bitrise
// config path` used to fall into: it printed the new config.yml path
// unconditionally, so `cat $(bitrise config path)` failed with "no such
// file" for anyone still being served by the predecessor CLI's config.yaml
// fallback.
func TestPathCmd_PrintsPredecessorPathWhileFallbackLive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bitrise"), 0o700))
	predecessorFile := filepath.Join(dir, "bitrise", "config.yaml")
	require.NoError(t, os.WriteFile(predecessorFile, []byte("theme: dark\n"), 0o600))

	cmd, out := newTestCmd(t, NewPathCommand())
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, predecessorFile+"\n", out.String())
}

func TestPathCmd_RejectsPositionalArgs(t *testing.T) {
	cmd := NewPathCommand()
	cmd.SetArgs([]string{"unexpected"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "positional args should be rejected before the command runs")
}

// newTestCmd wires cmd's output into a buffer for assertions. Every cli/config
// subcommand only touches the global config file, isolated via XDG_CONFIG_HOME
// in each test — no auth/API-client wiring is needed here, unlike cli/stack.
func newTestCmd(t *testing.T, cmd *cobra.Command) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	return cmd, &out
}
