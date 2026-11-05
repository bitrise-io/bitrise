package quarantinedtests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

const quarantinedTestJSON = `{
	"id":"qt-1","test_case_name":"test_login","test_suite_name":["AppTests","LoginTests"],"class_name":"LoginSpec",
	"quarantined_at":"2026-11-04T10:15:00.000Z","comment":"Flaky on CI",
	"filters":{"branches":["main","release/*"],"targets":null},
	"quarantined_from":{"number":42,"source":"build","slug":"build-1"},"quarantined_by":"jdoe"
}`

func TestList_Table(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"quarantined_tests":[` + quarantinedTestJSON + `]}`))
	})

	out, _, err := runCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/quarantined-tests", gotPath)
	for _, want := range []string{"ID", "TEST", "BRANCHES", "TARGETS", "QUARANTINED", "qt-1", "AppTests › LoginTests › LoginSpec › test_login", "main, release/*", "all", "by jdoe"} {
		assert.Contains(t, out, want)
	}
}

func TestList_LeavesOutClassNameTheNameStartsWith(t *testing.T) {
	const jestTest = `{"id":"qt-2","test_case_name":"Settings when saving shows a toast","class_name":"Settings when saving shows a toast",
		"test_suite_name":["Settings"],"quarantined_at":"2026-11-04T10:15:00.000Z","filters":{},"quarantined_from":{"source":"insights"}}`
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"quarantined_tests":[` + jestTest + `]}`))
	})

	out, _, err := runCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "Settings › Settings when saving shows a toast ")
	assert.Equal(t, 1, strings.Count(out, "when saving shows a toast"))

	out, _, err = runCmd(t, NewListCommand(), srv.URL, "--format", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"class_name": "Settings when saving shows a toast"`, "JSON keeps the class name as the API sent it")
}

func TestList_Empty(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"quarantined_tests":[]}`))
	})

	out, _, err := runCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "No quarantined tests found.\n", out)
}

func TestList_JSON(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"quarantined_tests":[` + quarantinedTestJSON + `]}`))
	})

	out, _, err := runCmd(t, NewListCommand(), srv.URL, "--format", "json")
	require.NoError(t, err)
	var got struct {
		Items []map[string]any `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	require.Len(t, got.Items, 1)
	assert.Equal(t, "qt-1", got.Items[0]["id"])
	assert.Equal(t, map[string]any{"source": "build", "id": "build-1", "number": float64(42)}, got.Items[0]["quarantined_from"])
}

func TestList_APIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error_msg":"Not Found"}`))
	})

	_, _, err := runCmd(t, NewListCommand(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing quarantined tests failed")
	assert.Contains(t, err.Error(), "Not Found")
}

// runCmd runs cmd against apiBaseURL with the app resolved from
// BITRISE_APP_ID, so tests see no app-name lookup, and returns its stdout and
// stderr.
func runCmd(t *testing.T, cmd *cobra.Command, apiBaseURL string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	t.Setenv("BITRISE_APP_ID", "my-app")
	t.Setenv("BITRISE_APP_SLUG", "")
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	// The root command sets these; without them cobra prints usage into out.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
