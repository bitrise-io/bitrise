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

func TestViewCmd_HappyPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"build-1","build_number":7,"status":1,"branch":"main","triggered_workflow":"primary"}}`))
	})

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Equal(t, "/apps/my-app/builds/build-1", gotPath)
	assert.Regexp(t, `Build #7`, out)
	assert.Regexp(t, `Status:\s+success`, out)
	assert.Regexp(t, `Branch:\s+main`, out)
	assert.Regexp(t, `Workflow:\s+primary`, out)
}

func TestViewCmd_EnvOnly_SkipsNameResolution(t *testing.T) {
	var gotPath string
	var titleLookupCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps" {
			titleLookupCalled = true
			_, _ = w.Write([]byte(`{"data":[],"paging":{}}`))
			return
		}
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"build-1","build_number":7,"status":1,"branch":"main","triggered_workflow":"primary"}}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv(cmdutil.EnvAppID, "env-app-slug")
	cmd := newTestViewCmd(t, srv.URL)
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Equal(t, "/apps/env-app-slug/builds/build-1", gotPath)
	assert.False(t, titleLookupCalled, "an ambient app slug (env/config) must not be resolved by name")
	assert.Regexp(t, `Build #7`, out)
}

func TestViewCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd := newTestViewCmd(t, "https://unused.test")
	_, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.EqualError(t, err, "--app is required")
}

func TestViewCmd_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	_, err := runViewCapture(t, cmd, []string{"missing"}, false, unusedBrowser(t))
	require.EqualError(t, err, `build "missing" not found`)
}

func TestViewCmd_Web_EnvOnly_SkipsAPICall(t *testing.T) {
	cmd := newTestViewCmd(t, "https://unused.test")
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")
	t.Setenv(cmdutil.EnvAppID, "env-app-slug")

	var gotURL string
	_, err := runViewCapture(t, cmd, []string{"build-1"}, true, func(url string) error {
		gotURL = url
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "https://app.bitrise.io/app/env-app-slug/build/build-1", gotURL)
}

func TestViewCmd_Web_ResolvesUserProvidedName(t *testing.T) {
	var gotTitle string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTitle = r.URL.Query().Get("title")
		_, _ = w.Write([]byte(`{"data":[{"slug":"app-123","title":"My App Name","owner":{}}],"paging":{}}`))
	}))
	t.Cleanup(srv.Close)

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "My App Name"))
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")

	var gotURL string
	_, err := runViewCapture(t, cmd, []string{"build-1"}, true, func(url string) error {
		gotURL = url
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "https://app.bitrise.io/app/app-123/build/build-1", gotURL)
	assert.Equal(t, "My App Name", gotTitle)
}

func TestViewCmd_RejectsWrongArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"a", "b"}} {
		cmd := NewViewCommand()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		assert.Error(t, cmd.Execute(), "args=%v should be rejected", args)
	}
}

// unusedBrowser fails the test if the browser opener is ever invoked, for
// cases where --web is off and no browser call is expected.
func unusedBrowser(t *testing.T) func(string) error {
	t.Helper()
	return func(url string) error {
		t.Fatalf("unexpected browser open: %s", url)
		return nil
	}
}

// runViewCapture calls runView and returns whatever it wrote to cmd's output.
func runViewCapture(t *testing.T, cmd *cobra.Command, args []string, web bool, openBrowser func(string) error) (string, error) {
	t.Helper()
	err := runView(cmd, args, web, openBrowser)
	buf, ok := cmd.OutOrStdout().(*bytes.Buffer)
	require.True(t, ok)
	return buf.String(), err
}

// newTestViewCmd builds a NewViewCommand() wired to apiBaseURL, with its
// output captured for runViewCapture to retrieve.
func newTestViewCmd(t *testing.T, apiBaseURL string) *cobra.Command {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewViewCommand()
	cmd.SetOut(&bytes.Buffer{})
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd
}
