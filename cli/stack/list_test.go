package stack

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
	internalstack "github.com/bitrise-io/bitrise/v2/internal/stack"
)

func TestListCmd_GlobalPath(t *testing.T) {
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
