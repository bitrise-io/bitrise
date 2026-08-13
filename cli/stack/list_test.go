package stack

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	internalstack "github.com/bitrise-io/bitrise/v2/internal/stack"
)

func TestListCmd_GlobalPath(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"linux-docker-android": {"title":"Linux Android","os":"linux","status":"stable"}}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/available-stacks", gotPath)
	assert.Contains(t, out.String(), "linux-docker-android")
}

func TestListCmd_WorkspaceScopedPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("workspace", "my-workspace"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/organizations/my-workspace/available-stacks", gotPath)
}

func TestListCmd_WorkspaceFromEnv(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv(cmdutil.EnvWorkspaceID, "env-ws")
	cmd, _ := newTestListCmd(t, srv.URL)
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/organizations/env-ws/available-stacks", gotPath)
	assert.Contains(t, errOut.String(), "Using default workspace: env-ws")
}

func TestListCmd_WorkspaceFromConfig(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestListCmdWithConfig(t, srv.URL, config.Config{DefaultWorkspaceID: "cfg-ws"})
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/organizations/cfg-ws/available-stacks", gotPath)
	assert.Contains(t, errOut.String(), "Using default workspace: cfg-ws")
}

func TestListCmd_WorkspaceFlagWinsOverDefault_NoBreadcrumb(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv(cmdutil.EnvWorkspaceID, "env-ws")
	cmd, _ := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("workspace", "flag-ws"))
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/organizations/flag-ws/available-stacks", gotPath)
	assert.NotContains(t, errOut.String(), "Using default workspace", "an explicit --workspace is not a default")
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "No stacks found.\n", out.String())
}

func TestPrintStacksTable(t *testing.T) {
	var buf bytes.Buffer
	err := printStacksTable(&buf, []internalstack.Stack{
		{ID: "a-stack", Title: "A", OS: "linux", Status: "stable"},
		{ID: "b-stack", Title: "B", OS: "osx", Status: "frozen", RemovalDate: "2027-01-01"},
	})
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "TITLE")
	assert.Contains(t, out, "a-stack")
	assert.Contains(t, out, "b-stack")
	assert.Contains(t, out, "2027-01-01")
}

func TestPrintStacksTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printStacksTable(&buf, nil))
	assert.Equal(t, "No stacks found.\n", buf.String())
}

func TestListCmd_PropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestListCmd(t, srv.URL)
	err := cmd.RunE(cmd, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing stacks failed")
	assert.Contains(t, err.Error(), "forbidden")
}

func TestListCmd_RejectsPositionalArgs(t *testing.T) {
	cmd := NewListCommand()
	cmd.SetArgs([]string{"unexpected"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "positional args should be rejected before the command runs")
}

// newTestListCmd builds a NewListCommand() wired to apiBaseURL, with its
// output captured in the returned buffer.
func newTestListCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewListCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}

// newTestListCmdWithConfig is newTestListCmd for tests that need extra
// resolved config keys (e.g. default_workspace_id) alongside the API base URL.
func newTestListCmdWithConfig(t *testing.T, apiBaseURL string, global config.Config) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd, out := newTestListCmd(t, apiBaseURL)
	global.APIBaseURL = apiBaseURL
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(config.Config{}, config.Config{}, global)))
	return cmd, out
}
