package config

import (
	"bytes"
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
