package build

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
)

func TestListCmd_PrintsTable(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/apps/my-app/builds", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"slug":"build-1","build_number":42,"status":1,"branch":"main","triggered_workflow":"primary"}]}`))
	})

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, out.String(), "#42")
	assert.Contains(t, out.String(), "main")
	assert.Contains(t, out.String(), "primary")
	assert.Contains(t, out.String(), "build-1")
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "No builds found.\n", out.String())
}

func TestListCmd_PrintsNextPageHint(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"build-1","build_number":1,"status":1}],"paging":{"next":"page-2"}}`))
	})

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("branch", "main"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, out.String(), "More results available")
	assert.Contains(t, out.String(), "bitrise build list --app my-app --branch main --cursor page-2")
}

func TestListCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestListCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, "--app is required")
}

func TestListCmd_AllFetchesEveryPage(t *testing.T) {
	pages := 0
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		pages++
		if r.URL.Query().Get("next") == "" {
			_, _ = w.Write([]byte(`{"data":[{"slug":"build-1","build_number":1,"status":1}],"paging":{"next":"page-2"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"slug":"build-2","build_number":2,"status":1}]}`))
	})

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("all", "true"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, 2, pages)
	assert.Contains(t, out.String(), "build-1")
	assert.Contains(t, out.String(), "build-2")
	assert.NotContains(t, out.String(), "More results available")
}

func TestListCmd_AllStopsOnRepeatedCursor(t *testing.T) {
	requests := 0
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"data":[{"slug":"build-1","build_number":1,"status":1}],"paging":{"next":"page-2"}}`))
	})

	cmd, _ := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("all", "true"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `pagination stalled: the API returned the cursor "page-2" twice`)
	assert.Equal(t, 2, requests, "must stop on the second repeat, not loop forever")
}

func TestListCmd_AllEmptyResultEmitsEmptyArray(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	cmd, out := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("all", "true"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, out.String(), `"items": []`, "an empty --all result must match the single-page path's non-nil items array")
}

func TestListCmd_RejectsAllWithCursor(t *testing.T) {
	cmd, _ := newTestListCmd(t, "https://unused.test")
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("all", "true"))
	require.NoError(t, cmd.Flags().Set("cursor", "x"))

	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, "--all and --cursor cannot be used together")
}

func TestListCmd_InvalidAfterValue(t *testing.T) {
	cmd, _ := newTestListCmd(t, "https://unused.test")
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("after", "not-a-date"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --after value")
}

func TestListCmd_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})

	cmd, _ := newTestListCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	err := cmd.RunE(cmd, nil)

	require.Error(t, err)
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

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
