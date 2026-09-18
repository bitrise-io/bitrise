package cmdutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// A bare command has a nil Context(), which must read as "no default set"
// rather than panicking.
func TestDefaultWorkspaceSlug_EmptyWhenNothingSet(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "")

	assert.Empty(t, DefaultWorkspaceSlug(&cobra.Command{}))
}

func TestDefaultWorkspaceSlug_FromConfig(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(
		config.Config{}, config.Config{}, config.Config{DefaultWorkspaceID: "cfg-ws"},
	)))

	assert.Equal(t, "cfg-ws", DefaultWorkspaceSlug(cmd))
}

func TestDefaultWorkspaceSlug_EnvWinsOverConfig(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "env-ws")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(
		config.Config{}, config.Config{}, config.Config{DefaultWorkspaceID: "cfg-ws"},
	)))

	assert.Equal(t, "env-ws", DefaultWorkspaceSlug(cmd))
}

func TestResolveWorkspaceID_FlagWins(t *testing.T) {
	cmd := newTestWorkspaceCmd(t)
	require.NoError(t, cmd.Flags().Set(FlagWorkspace, "flag-ws"))

	ws, err := ResolveWorkspaceID(cmd)
	require.NoError(t, err)
	assert.Equal(t, "flag-ws", ws)
}

func TestResolveWorkspaceID_FallsBackToDefaultWorkspaceSlug(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "env-ws")
	cmd := newTestWorkspaceCmd(t)

	ws, err := ResolveWorkspaceID(cmd)
	require.NoError(t, err)
	assert.Equal(t, "env-ws", ws)
}

func TestResolveWorkspaceID_SoleWorkspaceAutoDetect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"only-ws","name":"Only Workspace"}]}`))
	}))
	t.Cleanup(srv.Close)

	cmd, errOut := newTestWorkspaceCmdWithServer(t, srv.URL)

	ws, err := ResolveWorkspaceID(cmd)
	require.NoError(t, err)
	assert.Equal(t, "only-ws", ws)
	assert.Contains(t, errOut.String(), "Using your only workspace: Only Workspace (only-ws)")
	assert.Contains(t, errOut.String(), "Set it permanently to skip this lookup: bitrise config set default_workspace_id only-ws")
}

func TestResolveWorkspaceID_SoleWorkspaceBreadcrumbSuppressedByQuiet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"only-ws"}]}`))
	}))
	t.Cleanup(srv.Close)

	cmd, errOut := newTestWorkspaceCmdWithServer(t, srv.URL)
	require.NoError(t, cmd.PersistentFlags().Set(FlagQuiet, "true"))

	ws, err := ResolveWorkspaceID(cmd)
	require.NoError(t, err)
	assert.Equal(t, "only-ws", ws)
	assert.Empty(t, errOut.String())
}

func TestResolveWorkspaceID_SoleWorkspaceBreadcrumbSuppressedByFormat(t *testing.T) {
	for _, format := range []string{output.FormatJSON, output.FormatYML} {
		t.Run(format, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"data":[{"slug":"only-ws"}]}`))
			}))
			t.Cleanup(srv.Close)

			output.Format = format
			t.Cleanup(func() { output.Format = output.FormatRaw })

			cmd, errOut := newTestWorkspaceCmdWithServer(t, srv.URL)

			ws, err := ResolveWorkspaceID(cmd)
			require.NoError(t, err)
			assert.Equal(t, "only-ws", ws)
			assert.Empty(t, errOut.String())
		})
	}
}

func TestResolveWorkspaceID_AmbiguousWorkspaceFriendlyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"b-ws","name":"Bravo"},{"slug":"a-ws","name":"Alpha"}]}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestWorkspaceCmdWithServer(t, srv.URL)

	// Non-interactive with 2+ workspaces surfaces workspace.Sole's error
	// rather than prompting; its wording and ordering are pinned in
	// internal/workspace.
	_, err := ResolveWorkspaceID(cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple workspaces available")
}

func TestResolveWorkspaceID_NoWorkspaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestWorkspaceCmdWithServer(t, srv.URL)

	_, err := ResolveWorkspaceID(cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspaces found")
}

// newTestWorkspaceCmd builds a bare command with --workspace and the root
// --quiet flag registered, its stderr captured (never a terminal, so
// ResolveWorkspaceID never mistakes the test process for an interactive one).
func newTestWorkspaceCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String(FlagWorkspace, "", "workspace")
	cmd.PersistentFlags().Bool(FlagQuiet, false, "quiet")
	cmd.SetErr(&bytes.Buffer{})
	return cmd
}

// newTestWorkspaceCmdWithServer is newTestWorkspaceCmd wired to a stub
// /organizations server, with its captured stderr returned separately.
func newTestWorkspaceCmdWithServer(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv(EnvWorkspaceID, "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := newTestWorkspaceCmd(t)
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, errOut
}
