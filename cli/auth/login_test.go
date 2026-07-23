package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

func newTestCmd(t *testing.T, stdin string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	return cmd
}

func TestRunTokenLogin_SavesToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	require.NoError(t, runTokenLogin(newTestCmd(t, "bitpat_faketoken\n")))

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_faketoken", saved.Token)
	assert.False(t, saved.IsOAuthManaged())
}

func TestRunTokenLogin_EmptyTokenErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := runTokenLogin(newTestCmd(t, "\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is empty")
}
