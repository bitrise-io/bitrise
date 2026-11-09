package build

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestViewCmd_HappyPath(t *testing.T) {
	var gotBuildPath, gotSummaryPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/log/summary") {
			gotSummaryPath = r.URL.Path
			_, _ = w.Write([]byte(`{"agent_info":{"stack_image_version":"2-161-0"}}`))
			return
		}
		gotBuildPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"build-1","build_number":7,"status":1,"branch":"main","triggered_workflow":"primary"}}`))
	})

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Equal(t, "/apps/my-app/builds/build-1", gotBuildPath)
	assert.Equal(t, "/apps/my-app/builds/build-1/log/summary", gotSummaryPath)
	assert.Regexp(t, `Build #7`, out)
	assert.Regexp(t, `Status:\s+success`, out)
	assert.Regexp(t, `Branch:\s+main`, out)
	assert.Regexp(t, `Workflow:\s+primary`, out)
	assert.Regexp(t, `Image:\s+2-161-0`, out)
	assert.Empty(t, stderrOf(t, cmd))
}

func TestViewCmd_StackImageVersionInJSON(t *testing.T) {
	srv := newViewFakeServer(t, http.StatusOK, `{"agent_info":{"stack_image_version":"2-161-0"}}`)

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set(cmdutil.FormatKey, "json"))
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Contains(t, out, `"stack_image_version": "2-161-0"`)
}

// A build whose VM never started has no image version. That is normal, so
// the image line is left out and nothing is reported. The body is captured
// from a real build aborted as idle before it reached a worker: the summary
// exists, but agent_info is absent.
func TestViewCmd_VMNeverStarted_OmitsImageSilently(t *testing.T) {
	srv := newViewFakeServer(t, http.StatusOK, `{"app_id":"my-app","build_id":"build-1","has_build_environment_setup_logs":false,"execution":{"workflows":[]},"is_log_archived":false}`)

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Regexp(t, `Build #7`, out)
	assert.NotContains(t, out, "Image:")
	assert.Empty(t, stderrOf(t, cmd))
}

// The image version is supplementary: a failed summary fetch must not fail
// the command, but the user is told why the field is missing.
func TestViewCmd_LogSummaryError_WarnsAndPrintsBuild(t *testing.T) {
	srv := newViewFakeServer(t, http.StatusInternalServerError, `{"message":"boom"}`)

	cmd := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	out, err := runViewCapture(t, cmd, []string{"build-1"}, false, unusedBrowser(t))
	require.NoError(t, err)

	assert.Regexp(t, `Build #7`, out)
	assert.NotContains(t, out, "Image:")
	assert.Regexp(t, `warning: could not fetch stack image version: .*bitrise API 500: boom`, stderrOf(t, cmd))
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
		if strings.HasSuffix(r.URL.Path, "/log/summary") {
			w.WriteHeader(http.StatusNotFound)
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
	cmd.SetErr(&bytes.Buffer{})
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd
}

// newViewFakeServer serves build-1 for the build GET and answers the log
// summary GET with summaryStatus and summaryBody.
func newViewFakeServer(t *testing.T, summaryStatus int, summaryBody string) *httptest.Server {
	t.Helper()
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/log/summary") {
			w.WriteHeader(summaryStatus)
			_, _ = w.Write([]byte(summaryBody))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"slug":"build-1","build_number":7,"status":1}}`))
	})
}

// stderrOf returns what the command wrote to its error stream.
func stderrOf(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	buf, ok := cmd.ErrOrStderr().(*bytes.Buffer)
	require.True(t, ok)
	return buf.String()
}
