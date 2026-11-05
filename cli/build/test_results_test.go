package build

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestTestResultsCmd_HappyPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(testResultsJSON))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "/apps/my-app/builds/build-1/test-results", gotPath)
	assert.Equal(t, `LoginTests  4 tests, 1 failed, 1 errors, 0 skipped, 2.8s
  passed   LoginTests.testOK  120ms  suites: LoginTests
  failed   LoginTests.testWrongPassword  1.1s  suites: AppTests › LoginTests  identifier: LoginTests/testWrongPassword()
    expected error
    at line 12
  passed   LoginTests.testRetry  1.3s  flaky  2 runs
  SSO  1 tests, 0 failed, 1 errors, 0 skipped, 300ms
    error    SSOTests.testSAML  300ms
      crashed
Empty  0 tests, 0 failed, 0 errors, 0 skipped, 0s

Total: 4 tests, 1 failed, 1 errors, 0 skipped, 2.8s

Quarantining a test case needs its exact suites, name and class name, or its identifier. To look them up in this build:
  bitrise app quarantined-tests add --app my-app --build build-1 --name NAME
`, out.String())
}

func TestTestResultsCmd_SendsFilters(t *testing.T) {
	var gotQuery url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"test_suites": [{"name": "LoginTests", "test_count": 3, "failure_count": 1, "duration_in_ms": 2500,
  "test_cases": [{"name": "testWrongPassword", "class_name": "LoginTests", "status": "failed", "message": "boom", "duration_in_ms": 1100}],
  "test_suites": []}]}`))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "failed,error", "--suite", "LoginTests", "--class-name", "LoginTests",
		"--flaky=false", "--sort", "duration", "--order", "desc", "--limit", "20"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, url.Values{
		"status":     {"failed", "error"},
		"suite":      {"LoginTests"},
		"class_name": {"LoginTests"},
		"flaky":      {"false"},
		"sort_by":    {"duration"},
		"sort_order": {"desc"},
		"limit":      {"20"},
	}, gotQuery)
	assert.Contains(t, out.String(), "testWrongPassword")
	assert.NotContains(t, out.String(), "Total:")
	assert.Contains(t, out.String(), "\n1 matching test case.\n")
}

func TestTestResultsCmd_FlakyUnsetSendsNothing(t *testing.T) {
	var gotQuery url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(testResultsJSON))
	})

	cmd, _ := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())
	assert.Empty(t, gotQuery)

	cmd, _ = newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--flaky"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, url.Values{"flaky": {"true"}}, gotQuery)
}

func TestTestResultsCmd_FilterNoMatch(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"test_suites": []}`))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "skipped"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "No test cases match the filters.\n", out.String())
}

func TestTestResultsCmd_NoResults(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"test_suites": []}`))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "No test results found.\n", out.String())
}

func TestTestResultsCmd_JSON(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testResultsJSON))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--name", "retry", "--format", "json"})
	require.NoError(t, cmd.Execute())

	var got struct {
		TestSuites []struct {
			Name      string `json:"name"`
			TestCount int    `json:"test_count"`
			TestCases []struct {
				Name  string `json:"name"`
				Flaky bool   `json:"flaky"`
			} `json:"test_cases"`
		} `json:"test_suites"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got), out.String())
	require.Len(t, got.TestSuites, 1)
	assert.Equal(t, "LoginTests", got.TestSuites[0].Name)
	assert.Equal(t, 4, got.TestSuites[0].TestCount)
	require.Len(t, got.TestSuites[0].TestCases, 1)
	assert.Equal(t, "testRetry", got.TestSuites[0].TestCases[0].Name)
}

func TestTestResultsCmd_JSONKeepsSuitesAndIdentifier(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testResultsJSON))
	})

	cmd, out := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--format", "json"})
	require.NoError(t, cmd.Execute())

	var got struct {
		TestSuites []struct {
			TestCases []map[string]any `json:"test_cases"`
		} `json:"test_suites"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got), out.String())
	require.NotEmpty(t, got.TestSuites)
	require.Len(t, got.TestSuites[0].TestCases, 3)
	assert.NotContains(t, got.TestSuites[0].TestCases[2], "test_suites", "a case without suites leaves the key out")
	tc := got.TestSuites[0].TestCases[1]
	assert.Equal(t, []any{"AppTests", "LoginTests"}, tc["test_suites"])
	assert.Equal(t, "LoginTests/testWrongPassword()", tc["identifier"])
	assert.NotContains(t, tc, "properties", "only the identifier of the properties is exposed")
}

func TestTestResultsCmd_BuildNotFinished(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Test results are available once the build has finished"}`))
	})

	cmd, _ := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.EqualError(t, cmd.Execute(), "Test results are available once the build has finished")
}

func TestTestResultsCmd_UnknownStatus(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, "/apps/my-app/builds/build-1/test-results", r.URL.Path, "an invalid filter must fail before fetching")
		_, _ = w.Write([]byte(`{"data": []}`))
	})

	cmd, _ := newTestTestResultsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "broken"})
	require.ErrorContains(t, cmd.Execute(), `unknown test status "broken"`)
}

func TestTestResultsCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestTestResultsCmd(t, "https://unused.test")
	cmd.SetArgs([]string{"build-1"})
	require.EqualError(t, cmd.Execute(), "--app is required")
}

func TestTestResultsCmd_RejectsWrongArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"a", "b"}} {
		cmd := NewTestResultsCommand()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		assert.Error(t, cmd.Execute(), "args=%v should be rejected", args)
	}
}

func newTestTestResultsCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewTestResultsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}

const testResultsJSON = `{"test_suites": [
  {"name": "LoginTests", "test_count": 4, "failure_count": 1, "error_count": 1, "skipped_count": 0, "duration_in_ms": 2800,
   "test_cases": [
     {"name": "testOK", "class_name": "LoginTests", "status": "passed", "duration_in_ms": 120, "test_suites": ["LoginTests"]},
     {"name": "testWrongPassword", "class_name": "LoginTests", "status": "failed", "message": "expected error\nat line 12", "duration_in_ms": 1100,
      "test_suites": ["AppTests", "LoginTests"],
      "properties": [{"key": "configuration-hash", "value": "abc"}, {"key": "test-identifier", "value": "LoginTests/testWrongPassword()"}]},
     {"name": "testRetry", "class_name": "LoginTests", "status": "passed", "duration_in_ms": 1300, "flaky": true,
      "runs": [{"status": "failed", "message": "timeout", "duration_in_ms": 900}, {"status": "passed", "duration_in_ms": 400}]}
   ],
   "test_suites": [
     {"name": "SSO", "test_count": 1, "failure_count": 0, "error_count": 1, "skipped_count": 0, "duration_in_ms": 300,
      "test_cases": [{"name": "testSAML", "class_name": "SSOTests", "status": "error", "message": "crashed", "duration_in_ms": 300}],
      "test_suites": []}
   ]},
  {"name": "Empty", "test_count": 0, "failure_count": 0, "error_count": 0, "skipped_count": 0, "duration_in_ms": 0, "test_cases": [], "test_suites": []}
]}`
