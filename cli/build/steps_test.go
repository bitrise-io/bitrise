package build

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestStepsCmd_HappyPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "/apps/my-app/builds/build-1/log/summary", gotPath)
	assert.Regexp(t, `WORKFLOW\s+STEP\s+STATUS\s+DURATION`, out.String())
	assert.Regexp(t, `_setup\s+Git clone\s+success\s+<1s`, out.String())
	assert.Regexp(t, `primary\s+Tests\s+failed\s+1m45s`, out.String())
	assert.Regexp(t, `primary\s+save-cache\s+skipped\s*\n`, out.String(), "a skipped step has no duration, and without a title it falls back to its step ID")
	assert.Contains(t, out.String(), "Reasons:\n  primary › Tests\n    first line\n    second line\n")
}

func TestStepsCmd_StatusFilter(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "failed"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Tests")
	assert.NotContains(t, out.String(), "Git clone")
	assert.NotContains(t, out.String(), "save-cache")
	assert.Contains(t, out.String(), "first line")
}

func TestStepsCmd_StatusFilterNoMatch(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "preparation_failed"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "No steps with status \"preparation_failed\".\n", out.String())
}

func TestStepsCmd_JSON(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--status", "failed", "--format", "json"})
	require.NoError(t, cmd.Execute())

	var got struct {
		Items []struct {
			ID       string `json:"id"`
			Workflow string `json:"workflow"`
			StepID   string `json:"step_id"`
			Status   string `json:"status"`
			Errors   []struct {
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got), out.String())
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2", got.Items[0].ID)
	assert.Equal(t, "primary", got.Items[0].Workflow)
	assert.Equal(t, "script", got.Items[0].StepID)
	assert.Equal(t, "failed", got.Items[0].Status)
	require.Len(t, got.Items[0].Errors, 1)
	assert.Equal(t, "first line\nsecond line", got.Items[0].Errors[0].Message)
}

func TestStepsCmd_RunningBuildNotice(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": false, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "script", "status": "started", "started_at": "2026-10-20T11:56:50Z"}
]}]}}`))
	})

	for format, wantNotice := range map[string]bool{"raw": true, "json": false} {
		cmd, out := newTestStepsCmd(t, srv.URL)
		var errOut bytes.Buffer
		cmd.SetErr(&errOut)
		cmd.SetArgs([]string{"build-1", "--app", "my-app", "--format", format})
		require.NoError(t, cmd.Execute())

		if wantNotice {
			assert.Regexp(t, `primary\s+script\s+started`, out.String())
			assert.Contains(t, errOut.String(), "The step summary is not final yet")
		} else {
			assert.Contains(t, out.String(), `"complete": false`)
			assert.Empty(t, errOut.String())
		}
	}
}

func TestStepsCmd_Reasons(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": true, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "title": "Lint", "status": "skipped_with_run_if",
   "status_reason": "This Step was skipped, because its \"run_if\" expression evaluated to false.\nThe \"run_if\" expression was: false"},
  {"uuid": "s-2", "title": "Flaky tests", "status": "failed_skippable",
   "status_reason": "This Step failed, but it was marked as \"is_skippable\", so the build continued.",
   "errors": [{"code": 1, "message": "exit status 1"}]},
  {"uuid": "s-3", "title": "Tests", "status": "failed", "errors": [{"code": 1, "message": "exit status 2"}]},
  {"uuid": "s-4", "title": "Save cache", "status": "skipped",
   "status_reason": "This Step was skipped, because a previous Step failed, and this Step was not marked \"is_always_run\"."}
]}]}}`))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), `Reasons:
  primary › Lint
    This Step was skipped, because its "run_if" expression evaluated to false.
    The "run_if" expression was: false
  primary › Flaky tests
    This Step failed, but it was marked as "is_skippable", so the build continued.
    exit status 1
  primary › Tests
    exit status 2
`)
	assert.NotContains(t, out.String(), "previous Step failed", "a plain skipped step's reason is implied by the failure listed before it")
}

func TestStepsCmd_ErrorWithoutMessageShowsCode(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": true, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "script", "status": "failed", "errors": [{"code": 2, "message": ""}]}
]}]}}`))
	})

	cmd, out := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"build-1", "--app", "my-app"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "  primary › script\n    (no message, code 2)\n")
}

func TestStepsCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestStepsCmd(t, "https://unused.test")
	cmd.SetArgs([]string{"build-1"})
	require.EqualError(t, cmd.Execute(), "--app is required")
}

func TestStepsCmd_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _ := newTestStepsCmd(t, srv.URL)
	cmd.SetArgs([]string{"missing", "--app", "my-app"})
	require.EqualError(t, cmd.Execute(), `build "missing" not found`)
}

func TestStepsCmd_RejectsWrongArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"a", "b"}} {
		cmd := NewStepsCommand()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		assert.Error(t, cmd.Execute(), "args=%v should be rejected", args)
	}
}

func newTestStepsCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewStepsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}

const stepsSummaryJSON = `{
  "is_log_archived": true,
  "execution": {
    "workflows": [
      {"workflow_id": "_setup", "steps": [
        {"uuid": "s-1", "step_id": "git-clone", "title": "Git clone", "status": "success",
         "started_at": "2026-10-19T13:25:35Z", "finished_at": "2026-10-19T13:25:35.200Z"}
      ]},
      {"workflow_id": "primary", "steps": [
        {"uuid": "s-2", "step_id": "script", "title": "Tests", "status": "failed",
         "started_at": "2026-10-19T13:26:48Z", "finished_at": "2026-10-19T13:28:33Z",
         "errors": [{"code": 1, "message": "first line\nsecond line"}]},
        {"uuid": "s-3", "step_id": "save-cache", "status": "skipped",
         "started_at": "2026-10-19T13:28:33Z", "finished_at": "2026-10-19T13:28:33.300Z"}
      ]}
    ]
  }
}`
