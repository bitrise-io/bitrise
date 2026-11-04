package build

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
)

func TestService_List_PathAndMapping(t *testing.T) {
	var gotPath, gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{
  "data": [
    {"slug":"b-1","build_number":42,"status":1,"status_text":"success","branch":"main","triggered_workflow":"primary",
     "triggered_at":"2026-05-06T10:00:00Z","finished_at":"2026-05-06T10:05:00Z"},
    {"slug":"b-2","build_number":41,"status":0,"branch":"feature","triggered_workflow":"primary","triggered_at":"2026-05-06T09:00:00Z"}
  ],
  "paging": {"next": "next-cur"}
}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	res, err := svc.List(context.Background(), ListOptions{
		AppSlug:  "my-app",
		Branch:   "main",
		Workflow: "primary",
		Status:   "success",
		Limit:    50,
	})
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds", gotPath)
	require.Len(t, res.Items, 2)
	assert.Equal(t, "success", res.Items[0].Status)
	assert.Equal(t, "in-progress", res.Items[1].Status)
	assert.NotNil(t, res.Items[0].FinishedAt)
	assert.Nil(t, res.Items[1].FinishedAt)
	assert.Equal(t, "my-app", res.Items[0].AppSlug)
	assert.Equal(t, "next-cur", res.NextCursor)
	assert.Contains(t, gotQuery, "status=1")
}

func TestService_List_StatusInProgressMapsToZero(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.List(context.Background(), ListOptions{AppSlug: "my-app", Status: "in-progress"})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "status=0")
}

func TestService_List_NewFiltersPassedThrough(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	after := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	isPipeline := true

	_, err := svc.List(context.Background(), ListOptions{
		AppSlug:          "my-app",
		SortBy:           "running_first",
		CommitMessage:    "fix bug",
		TriggerEventType: "push",
		PullRequestID:    42,
		BuildNumber:      99,
		After:            &after,
		Before:           &before,
		IsPipelineBuild:  &isPipeline,
	})
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "running_first", q.Get("sort_by"))
	assert.Equal(t, "fix bug", q.Get("commit_message"))
	assert.Equal(t, "push", q.Get("trigger_event_type"))
	assert.Equal(t, "42", q.Get("pull_request_id"))
	assert.Equal(t, "99", q.Get("build_number"))
	assert.Equal(t, strconv.FormatInt(after.Unix(), 10), q.Get("after"))
	assert.Equal(t, strconv.FormatInt(before.Unix(), 10), q.Get("before"))
	assert.Equal(t, "true", q.Get("is_pipeline_build"))
}

func TestService_List_RejectsUnknownStatus(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.List(context.Background(), ListOptions{AppSlug: "my-app", Status: "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown build status")
}

func TestService_View_PathAndMapping(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"b-xyz","build_number":7,"status":2,"status_text":"failed","triggered_workflow":"deploy","branch":"feature/x"}}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.View(context.Background(), "my-app", "b-xyz")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds/b-xyz", gotPath)
	assert.Equal(t, "b-xyz", got.Slug)
	assert.Equal(t, "failed", got.Status)
	assert.Equal(t, "deploy", got.Workflow)
	assert.Equal(t, "my-app", got.AppSlug)
}

func TestService_View_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.View(context.Background(), "my-app", "missing")
	require.EqualError(t, err, `build "missing" not found`)
}

func TestService_Trigger_BodyAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
  "build_slug": "new-1",
  "build_number": 100,
  "build_url": "https://app.bitrise.io/build/new-1",
  "triggered_workflow": "deploy",
  "message": "Build triggered"
}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Trigger(context.Background(), TriggerRequest{
		AppSlug:    "my-app",
		Workflow:   "deploy",
		Branch:     "main",
		CommitHash: "abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds", gotPath)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	hi, _ := sent["hook_info"].(map[string]any)
	assert.Equal(t, "bitrise", hi["type"])
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, "deploy", bp["workflow_id"])
	assert.Equal(t, "main", bp["branch"])
	assert.Equal(t, "abc", bp["commit_hash"])

	assert.Equal(t, "new-1", got.Slug)
	assert.Equal(t, 100, got.BuildNumber)
	assert.NotEmpty(t, got.BuildURL)
	assert.Equal(t, "in-progress", got.Status)
	assert.Equal(t, "my-app", got.AppSlug)
	assert.Equal(t, "deploy", got.Workflow)
	assert.Equal(t, "main", got.Branch)
}

func TestService_Trigger_PrefersResultsArray(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// Modern API returns "results"; deprecated top-level fields differ.
		_, _ = w.Write([]byte(`{
  "build_slug": "deprecated-slug",
  "build_url": "deprecated-url",
  "results": [
    {"build_slug": "result-slug", "build_number": 99, "build_url": "result-url", "triggered_workflow": "deploy"}
  ]
}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Trigger(context.Background(), TriggerRequest{AppSlug: "a", Workflow: "deploy"})
	require.NoError(t, err)
	assert.Equal(t, "result-slug", got.Slug)
	assert.Equal(t, "result-url", got.BuildURL)
	assert.Equal(t, 99, got.BuildNumber)
}

func TestService_Trigger_RequiresAppSlug(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Trigger(context.Background(), TriggerRequest{Workflow: "x"})
	require.Error(t, err)
}

func TestService_Trigger_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Trigger(context.Background(), TriggerRequest{AppSlug: "missing-app", Workflow: "x"})
	require.EqualError(t, err, `app "missing-app" not found`)
}

func TestService_Trigger_PipelineFields(t *testing.T) {
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"p-1","build_number":5,"build_url":"https://app.bitrise.io/build/p-1"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Trigger(context.Background(), TriggerRequest{
		AppSlug:    "my-app",
		Pipeline:   "my-pipeline",
		Branch:     "main",
		BranchDest: "release",
		Tag:        "v1.0.0",
	})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, "my-pipeline", bp["pipeline_id"])
	assert.Equal(t, "release", bp["branch_dest"])
	assert.Equal(t, "v1.0.0", bp["tag"])
	assert.Nil(t, bp["workflow_id"])

	assert.Equal(t, "v1.0.0", got.Tag)
	assert.Equal(t, "main", got.Branch)
}

func TestService_Trigger_EnvsAndPriorityAndPR(t *testing.T) {
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"e-1","build_number":10}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	priority := 1
	_, err := svc.Trigger(context.Background(), TriggerRequest{
		AppSlug:       "my-app",
		Workflow:      "primary",
		PullRequestID: 42,
		Priority:      &priority,
		Environments: []TriggerEnv{
			{Key: "MY_VAR", Value: "hello", IsExpand: true},
			{Key: "RAW", Value: "$5", IsExpand: false},
		},
	})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, float64(42), bp["pull_request_id"])
	assert.Equal(t, float64(1), bp["priority"])
	envs, _ := bp["environments"].([]any)
	require.Len(t, envs, 2)
	env, _ := envs[0].(map[string]any)
	assert.Equal(t, "MY_VAR", env["mapped_to"])
	assert.Equal(t, "hello", env["value"])
	assert.Equal(t, true, env["is_expand"])
	raw, _ := envs[1].(map[string]any)
	assert.Equal(t, "RAW", raw["mapped_to"])
	assert.Equal(t, false, raw["is_expand"])
}

func TestService_Trigger_PriorityZeroIsSent(t *testing.T) {
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"z-1","build_number":1}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	zero := 0
	_, err := svc.Trigger(context.Background(), TriggerRequest{AppSlug: "my-app", Workflow: "primary", Priority: &zero})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	bp, _ := sent["build_params"].(map[string]any)
	got, ok := bp["priority"]
	require.True(t, ok, "priority 0 must be present in the body")
	assert.Equal(t, float64(0), got)
}

func TestService_Trigger_MachineFields(t *testing.T) {
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"m-1","build_number":1}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Trigger(context.Background(), TriggerRequest{
		AppSlug:       "my-app",
		Workflow:      "primary",
		Stack:         "osx-xcode-16.0.x",
		MachineTypeID: "g2-m1.4core",
		LicensePoolID: "pool-1",
	})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, "osx-xcode-16.0.x", bp["stack"])
	assert.Equal(t, "g2-m1.4core", bp["machine_type_id"])
	assert.Equal(t, "pool-1", bp["license_pool_id"])
	_, hasPriority := bp["priority"]
	assert.False(t, hasPriority, "priority must be omitted when not given")
}

func TestService_Log_Streams(t *testing.T) {
	rawSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("LOG OUTPUT"))
	}))
	t.Cleanup(rawSrv.Close)

	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_archived":          true,
			"expiring_raw_log_url": rawSrv.URL,
		})
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	complete, err := svc.Log(context.Background(), "my-app", "b-1", &buf)
	require.NoError(t, err)
	assert.True(t, complete)
	assert.Contains(t, buf.String(), "LOG OUTPUT")
}

func TestService_Log_RequiresSlugs(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Log(context.Background(), "", "b", io.Discard)
	assert.Error(t, err)
	_, err = svc.Log(context.Background(), "a", "", io.Discard)
	assert.Error(t, err)
}

func TestService_Steps_FlattensWorkflowsInOrder(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds/b-1/log/summary", gotPath)
	require.Len(t, got.Items, 3)
	assert.Equal(t, []string{"s-1", "s-2", "s-3"}, []string{got.Items[0].ID, got.Items[1].ID, got.Items[2].ID})
	assert.Equal(t, "_setup", got.Items[0].Workflow)
	assert.Equal(t, "primary", got.Items[1].Workflow)

	failed := got.Items[1]
	assert.Equal(t, "script", failed.StepID)
	assert.Equal(t, "Tests", failed.Title)
	assert.Equal(t, "failed", failed.Status)
	assert.Equal(t, []StepError{{Code: 1, Message: "exit status 1"}}, failed.Errors)
	require.NotNil(t, failed.StartedAt)
	assert.Equal(t, time.Date(2026, 10, 19, 13, 26, 48, 0, time.UTC), *failed.StartedAt)

	assert.True(t, got.Complete)

	skipped := got.Items[2]
	assert.Equal(t, "skipped", skipped.Status)
	assert.Contains(t, skipped.StatusReason, "previous Step failed")
	assert.Nil(t, skipped.StartedAt)
	assert.Nil(t, skipped.FinishedAt)
}

func TestService_Steps_SendsFilters(t *testing.T) {
	var gotQuery url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"is_log_archived": true, "execution": {"workflows": []}}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	tests := []struct {
		status, stepID string
		want           url.Values
	}{
		{"", "", url.Values{}},
		{"failed", "", url.Values{"status": {"failed", "preparation_failed", "aborted_with_custom_timeout", "aborted_with_no_output"}}},
		{"failed_skippable", "", url.Values{"status": {"failed_skippable"}}},
		{"", "xcode-test", url.Values{"step_id": {"xcode-test"}}},
		{"skipped", "script", url.Values{"status": {"skipped"}, "step_id": {"script"}}},
	}
	for _, tt := range tests {
		got, err := svc.Steps(context.Background(), "my-app", "b-1", tt.status, tt.stepID)
		require.NoError(t, err)
		assert.Equal(t, tt.want, gotQuery, "status %q, step ID %q", tt.status, tt.stepID)
		assert.NotNil(t, got.Items, "an empty result must render as [] not null")
	}
}

func TestService_Steps_IncompleteWhileLogNotArchived(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": false, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "git-clone", "status": "success"},
  {"uuid": "s-2", "step_id": "script", "status": "started", "started_at": "2026-10-20T11:56:50Z"}
]}]}}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
	require.NoError(t, err)
	assert.False(t, got.Complete)
	require.Len(t, got.Items, 2)
	assert.Equal(t, "started", got.Items[1].Status)
	assert.Nil(t, got.Items[1].FinishedAt)
}

func TestService_Steps_RequiresSlugs(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))

	_, err := svc.Steps(context.Background(), "", "b-1", "", "")
	require.EqualError(t, err, "app ID is required")
	_, err = svc.Steps(context.Background(), "my-app", "", "", "")
	require.EqualError(t, err, "build ID is required")
}

func TestService_Steps_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "missing", "", "")
	require.EqualError(t, err, `build "missing" not found`)
}

func TestService_Steps_NoSummaryYetForRunningBuild(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusOK, `{"data":{"slug":"b-1","status":0}}`)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
	require.NoError(t, err)
	assert.False(t, got.Complete)
	assert.NotNil(t, got.Items)
	assert.Empty(t, got.Items)
}

func TestService_Steps_NoSummaryForAbortedBuild(t *testing.T) {
	for _, status := range []int{3, 4} {
		srv := newSummaryNotFoundServer(t, http.StatusOK, fmt.Sprintf(`{"data":{"slug":"b-1","status":%d}}`, status))
		svc := NewService(newAPIClient(t, srv.URL))

		got, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
		require.NoError(t, err, "status %d", status)
		assert.True(t, got.Complete, "a build aborted before any step ran has its final, empty list")
		assert.NotNil(t, got.Items)
		assert.Empty(t, got.Items)
	}
}

func TestService_Steps_NoSummaryForFinishedBuild(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusOK, `{"data":{"slug":"b-1","status":2}}`)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
	require.EqualError(t, err, `build "b-1" has no step summary`)
}

func TestService_Steps_NoSummaryAndBuildLookupFails(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusInternalServerError, `{"message":"boom"}`)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "", "")
	require.ErrorContains(t, err, `no step summary for build "b-1", and looking up the build failed`)
}

func TestService_Steps_RejectsUnknownStatus(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "aborted-with-success", "")
	require.ErrorContains(t, err, `unknown step status "aborted-with-success" (expected: not_started, started, success`)
}

func TestStep_RunTime(t *testing.T) {
	start := time.Date(2026, 10, 19, 13, 26, 48, 0, time.UTC)
	end := start.Add(90 * time.Second)

	d, ok := Step{Status: "failed", StartedAt: &start, FinishedAt: &end}.RunTime()
	assert.True(t, ok)
	assert.Equal(t, 90*time.Second, d)

	_, ok = Step{Status: "skipped_with_run_if", StartedAt: &start, FinishedAt: &end}.RunTime()
	assert.False(t, ok, "a skipped step's timestamps don't mean it ran")
	_, ok = Step{Status: "started", StartedAt: &start}.RunTime()
	assert.False(t, ok)
}

func TestStep_FailedBuild(t *testing.T) {
	assert.True(t, Step{Status: "aborted_with_no_output"}.FailedBuild())
	assert.False(t, Step{Status: "failed_skippable"}.FailedBuild())
}

func TestStep_SkippedAfterFailure(t *testing.T) {
	assert.True(t, Step{Status: "skipped", StatusReason: `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`}.SkippedAfterFailure())
	assert.False(t, Step{Status: "skipped", StatusReason: "This Step was skipped, because the build was aborted."}.SkippedAfterFailure())
	assert.False(t, Step{Status: "skipped_with_run_if", StatusReason: "because a previous Step failed"}.SkippedAfterFailure())
}

func TestStepsResult_JSONShape(t *testing.T) {
	started := time.Date(2026, 10, 19, 13, 26, 48, 0, time.UTC)
	res := StepsResult{Items: []Step{{
		ID: "s-2", Workflow: "primary", StepID: "script", Title: "Tests", Version: "1.2.1",
		Status: "aborted_with_custom_timeout", StartedAt: &started, Errors: []StepError{{Code: 0, Message: "This Step timed out after 1m."}},
	}}, Complete: true}

	b, err := json.Marshal(res)
	require.NoError(t, err)
	assert.JSONEq(t, `{"complete":true,"items":[{"id":"s-2","workflow":"primary","step_id":"script","title":"Tests","version":"1.2.1",
		"status":"aborted_with_custom_timeout","started_at":"2026-10-19T13:26:48Z","errors":[{"code":0,"message":"This Step timed out after 1m."}]}]}`, string(b))

	y, err := yaml.Marshal(res)
	require.NoError(t, err)
	assert.Contains(t, string(y), "step_id: script")
	assert.Contains(t, string(y), "started_at:")
	assert.Contains(t, string(y), "code: 0")
	assert.Contains(t, string(y), "complete: true")
}

func TestService_StepLogs_FetchesLogOfEachMatchingStep(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "failed", "", nil, 0)
	require.NoError(t, err)
	assert.True(t, got.Complete)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2", got.Items[0].ID)
	assert.Equal(t, "primary", got.Items[0].Workflow)
	assert.Equal(t, "s-2 one\ns-2 two\ns-2 three\n", got.Items[0].Log)
}

func TestService_StepLogs_FiltersByStepIDAndStatus(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "", []string{"s-3", "s-1"}, 0)
	require.NoError(t, err)
	require.Len(t, got.Items, 2)
	assert.Equal(t, "s-1", got.Items[0].ID, "steps keep execution order, not the order of the IDs")
	assert.Equal(t, "s-3", got.Items[1].ID)
	assert.Empty(t, got.Items[1].Log, "a step that logged nothing has an empty log")

	got, err = svc.StepLogs(context.Background(), "my-app", "b-1", "failed", "", []string{"s-1"}, 0)
	require.NoError(t, err)
	assert.Empty(t, got.Items)
	assert.NotNil(t, got.Items, "an empty result must render as [] not null")
}

func TestService_StepLogs_StepIDFilter(t *testing.T) {
	var gotQuery url.Values
	steps := newStepLogsServer(t)
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/log/summary") {
			gotQuery = r.URL.Query()
		}
		steps.Config.Handler.ServeHTTP(w, r)
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "script", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, url.Values{"step_id": {"script"}}, gotQuery, "without step run UUIDs the server filters")

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "script", []string{"s-1"}, 0)
	require.NoError(t, err, "s-1 is git-clone: the server must not drop it, or it would look missing")
	assert.Empty(t, gotQuery)
	assert.Empty(t, got.Items)
}

func TestService_StepLogs_Tail(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "", []string{"s-2"}, 2)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2 two\ns-2 three\n", got.Items[0].Log)

	_, err = svc.StepLogs(context.Background(), "my-app", "b-1", "", "", nil, -1)
	require.Error(t, err)
}

func TestService_StepLogs_UnknownStepID(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "", []string{"s-2", "nope"}, 0)
	require.EqualError(t, err, `step "nope" not found in build "b-1"`)
}

func TestService_StepLogs_NoLogsUntilArchived(t *testing.T) {
	var stepLogCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/log/steps/") {
			stepLogCalls.Add(1)
		}
		_, _ = w.Write([]byte(`{"is_log_archived": false, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "script", "status": "started", "started_at": "2026-10-20T11:56:50Z"}
]}]}}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", "", []string{"s-1", "not-started-yet"}, 0)
	require.NoError(t, err)
	assert.False(t, got.Complete)
	require.Len(t, got.Items, 1)
	assert.Empty(t, got.Items[0].Log)
	assert.Zero(t, stepLogCalls.Load(), "step logs are served only from the archive")
}

func TestService_WaitForLogArchive(t *testing.T) {
	var calls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		archived := calls.Add(1) >= 3
		_, _ = fmt.Fprintf(w, `{"is_log_archived": %t}`, archived)
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var waiting int
	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Millisecond, func() { waiting++ }))
	assert.Equal(t, int32(3), calls.Load())
	assert.Equal(t, 1, waiting, "waiting is called once, when the first poll finds the log not archived")
}

func TestService_WaitForLogArchive_AlreadyArchived(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": true}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Hour, func() { t.Error("waiting must not be called") }))
}

func TestService_WaitForLogArchive_GivesUp(t *testing.T) {
	timeout := logArchiveTimeout
	logArchiveTimeout = 20 * time.Millisecond
	t.Cleanup(func() { logArchiveTimeout = timeout })

	var calls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"is_log_archived": false}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Millisecond, nil))
	assert.Greater(t, calls.Load(), int32(1))

	start := time.Now()
	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Hour, nil))
	assert.Less(t, time.Since(start), 5*time.Second, "an interval longer than the timeout must not outlast it")
}

func TestService_WaitForLogArchive_SlowResponseGivesUp(t *testing.T) {
	timeout := logArchiveTimeout
	logArchiveTimeout = 50 * time.Millisecond
	t.Cleanup(func() { logArchiveTimeout = timeout })

	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
			_, _ = w.Write([]byte(`{"is_log_archived": false}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	start := time.Now()
	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Millisecond, nil), "reaching the timeout mid-request is giving up, not an error")
	assert.Less(t, time.Since(start), 5*time.Second, "a slow response must not outlast the timeout")
}

func TestService_WaitForLogArchive_CallerCancel(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": false}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	ctx, cancel := context.WithCancel(context.Background())
	err := svc.WaitForLogArchive(ctx, "my-app", "b-1", time.Hour, cancel)
	require.ErrorIs(t, err, context.Canceled)
}

func TestService_WaitForLogArchive_ReturnsWithoutSummary(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	require.NoError(t, svc.WaitForLogArchive(context.Background(), "my-app", "b-1", time.Hour, nil))
}

func TestService_LogText(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"one\ntwo\n","position":0},{"chunk":"three\n","position":1}]}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.LogText(context.Background(), "my-app", "b-1", 0)
	require.NoError(t, err)
	assert.Equal(t, LogResult{Log: "one\ntwo\nthree\n", Complete: false}, got)

	got, err = svc.LogText(context.Background(), "my-app", "b-1", 2)
	require.NoError(t, err)
	assert.Equal(t, "two\nthree\n", got.Log)
}

func TestLogBuffer(t *testing.T) {
	for _, tt := range []struct {
		in   string
		n    int
		want string
	}{
		{"a\nb\nc\n", 2, "b\nc\n"},
		{"a\nb\nc", 2, "b\nc"},
		{"a\nb\nc", 1, "c"},
		{"a\nb\nc\n", 3, "a\nb\nc\n"},
		{"a\nb\nc\n", 10, "a\nb\nc\n"},
		{"a\nb\nc\n", 0, "a\nb\nc\n"},
		{"a\n\n\n", 2, "\n\n"},
		{"", 5, ""},
	} {
		// Every split point, since a line can span writes.
		for i := range len(tt.in) + 1 {
			buf := newLogBuffer(tt.n)
			_, _ = io.WriteString(buf, tt.in[:i])
			_, _ = io.WriteString(buf, tt.in[i:])
			assert.Equal(t, tt.want, buf.String(), "newLogBuffer(%d) given %q split at %d", tt.n, tt.in, i)
		}
	}
}

func TestStepLogsResult_Shape(t *testing.T) {
	res := StepLogsResult{Items: []StepLog{{Step: Step{ID: "s-2", Workflow: "primary", StepID: "script", Status: "failed"}, Log: "boom\n"}}, Complete: true}

	b, err := json.Marshal(res)
	require.NoError(t, err)
	assert.JSONEq(t, `{"complete":true,"items":[{"id":"s-2","workflow":"primary","step_id":"script","status":"failed","log":"boom\n"}]}`, string(b))

	y, err := yaml.Marshal(res)
	require.NoError(t, err)
	assert.Contains(t, string(y), "- id: s-2\n")
	assert.Contains(t, string(y), "step_id: script")
	assert.Contains(t, string(y), "log: |")
}

func TestService_Abort_BodyAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Abort(context.Background(), AbortRequest{
		AppSlug:          "my-app",
		BuildSlug:        "b-1",
		Reason:           "no longer needed",
		AbortWithSuccess: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds/b-1/abort", gotPath)
	assert.Equal(t, AbortResult{AppSlug: "my-app", BuildSlug: "b-1", Status: "ok"}, got)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "no longer needed", sent["abort_reason"])
	assert.Equal(t, true, sent["abort_with_success"])
}

func TestService_Abort_RequiresSlugs(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Abort(context.Background(), AbortRequest{BuildSlug: "b"})
	assert.Error(t, err)
	_, err = svc.Abort(context.Background(), AbortRequest{AppSlug: "a"})
	assert.Error(t, err)
}

func TestService_Abort_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Abort(context.Background(), AbortRequest{AppSlug: "my-app", BuildSlug: "missing"})
	require.EqualError(t, err, `build "missing" not found`)
}

func TestService_Rebuild_BodyAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"new-build","build_number":8,"build_url":"https://app.bitrise.io/build/new-build","triggered_workflow":"deploy"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	b, err := svc.Rebuild(context.Background(), RebuildRequest{AppSlug: "my-app", BuildSlug: "src-build", RemoteAccess: true})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds/src-build/rebuild", gotPath)
	assert.JSONEq(t, `{"is_remote":true}`, string(gotBody))
	assert.Equal(t, "new-build", b.Slug)
	assert.Equal(t, "my-app", b.AppSlug)
	assert.Equal(t, 8, b.BuildNumber)
	assert.Equal(t, "deploy", b.Workflow)
	assert.Equal(t, "https://app.bitrise.io/build/new-build", b.BuildURL)
}

func TestService_Rebuild_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Rebuild(context.Background(), RebuildRequest{AppSlug: "my-app", BuildSlug: "missing"})
	require.EqualError(t, err, `build "missing" not found`)
}

func TestService_Rebuild_PassesServerRejectionThrough(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_msg":"Build not finished yet, can't rebuild it."}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Rebuild(context.Background(), RebuildRequest{AppSlug: "my-app", BuildSlug: "running"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Build not finished yet, can't rebuild it.")
}

func TestService_Rebuild_RequiresSlugs(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {
		t.Error("no request expected")
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Rebuild(context.Background(), RebuildRequest{BuildSlug: "b"})
	require.EqualError(t, err, "app ID is required")
	_, err = svc.Rebuild(context.Background(), RebuildRequest{AppSlug: "my-app"})
	require.EqualError(t, err, "build ID is required")
}

func TestService_NilClientFails(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.List(context.Background(), ListOptions{AppSlug: "x"})
	assert.Error(t, err)
	_, err = svc.View(context.Background(), "a", "b")
	assert.Error(t, err)
	_, err = svc.Trigger(context.Background(), TriggerRequest{AppSlug: "a"})
	assert.Error(t, err)
	_, err = svc.Log(context.Background(), "a", "b", io.Discard)
	assert.Error(t, err)
	_, err = svc.Abort(context.Background(), AbortRequest{AppSlug: "a", BuildSlug: "b"})
	assert.Error(t, err)
}

func TestStatusString(t *testing.T) {
	cases := map[int]string{
		0: "in-progress",
		1: "success",
		2: "failed",
		3: "aborted",
		4: "aborted-with-success",
		9: "9", // unknown — passthrough as integer string
	}
	for in, want := range cases {
		assert.Equal(t, want, statusString(in))
	}
}

func TestParseStatusFilter(t *testing.T) {
	got, err := parseStatusFilter("")
	require.NoError(t, err)
	assert.Nil(t, got)

	got, err = parseStatusFilter("success")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, *got)

	got, err = parseStatusFilter("in-progress")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 0, *got)

	_, err = parseStatusFilter("BOGUS")
	assert.Error(t, err)
}

func TestService_TestResults_MapsTree(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(testResultsJSON))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.TestResults(context.Background(), "my-app", "b-1", TestResultsFilter{})
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds/b-1/test-results", gotPath)
	require.Len(t, got.TestSuites, 2)

	login := got.TestSuites[0]
	assert.Equal(t, "LoginTests", login.Name)
	assert.Equal(t, 3, login.TestCount)
	assert.Equal(t, 1, login.FailureCount)
	assert.Equal(t, int64(2500), login.DurationMS)
	require.Len(t, login.TestCases, 3)
	failed := login.TestCases[1]
	assert.Equal(t, "testWrongPassword", failed.Name)
	assert.Equal(t, "LoginTests", failed.ClassName)
	assert.Equal(t, "failed", failed.Status)
	assert.Equal(t, "expected error\nat line 12", failed.Message)
	flaky := login.TestCases[2]
	assert.True(t, flaky.Flaky)
	assert.Equal(t, []TestRun{{Status: "failed", Message: "timeout", DurationMS: 900}, {Status: "passed", DurationMS: 400}}, flaky.Runs)

	require.Len(t, login.TestSuites, 1)
	assert.Equal(t, "SSO", login.TestSuites[0].Name)
	require.Len(t, login.TestSuites[0].TestCases, 1)

	empty := got.TestSuites[1]
	assert.Equal(t, "Empty", empty.Name)
	assert.NotNil(t, empty.TestCases, "an empty suite must render as [] not null")
	assert.NotNil(t, empty.TestSuites)
}

func TestService_TestResults_SendsFilters(t *testing.T) {
	var gotQuery url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"test_suites": []}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))
	flaky, notFlaky := true, false

	tests := []struct {
		name   string
		filter TestResultsFilter
		want   url.Values
	}{
		{"none", TestResultsFilter{}, url.Values{}},
		{"statuses", TestResultsFilter{Statuses: []string{"failed", "error"}}, url.Values{"status": {"failed", "error"}}},
		{"suite and class name", TestResultsFilter{Suite: "LoginTests", ClassName: "SSOTests"}, url.Values{"suite": {"LoginTests"}, "class_name": {"SSOTests"}}},
		{"flaky", TestResultsFilter{Flaky: &flaky}, url.Values{"flaky": {"true"}}},
		{"not flaky", TestResultsFilter{Flaky: &notFlaky}, url.Values{"flaky": {"false"}}},
		{"sort", TestResultsFilter{SortBy: "duration", SortOrder: "desc"}, url.Values{"sort_by": {"duration"}, "sort_order": {"desc"}}},
		{"limit", TestResultsFilter{Limit: 20}, url.Values{"limit": {"20"}}},
		{"name is matched on the client", TestResultsFilter{Name: "sso"}, url.Values{}},
		{"limit with name is applied on the client", TestResultsFilter{Name: "sso", Limit: 20, Statuses: []string{"error"}}, url.Values{"status": {"error"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.TestResults(context.Background(), "my-app", "b-1", tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.want, gotQuery)
			assert.NotNil(t, got.TestSuites, "an empty result must render as [] not null")
		})
	}
}

func TestService_TestResults_NameFilter(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testResultsJSON))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	tests := []struct {
		name   string
		filter TestResultsFilter
		want   []string
	}{
		{"matches class name case-insensitively", TestResultsFilter{Name: "ssotests"}, []string{"SSO/testSAML"}},
		{"matches name", TestResultsFilter{Name: "PASSWORD"}, []string{"LoginTests/testWrongPassword"}},
		{"limit applies after it", TestResultsFilter{Name: "test", Limit: 2}, []string{"LoginTests/testOK", "LoginTests/testWrongPassword"}},
		{"limit counts a suite's own cases before its child suites'", TestResultsFilter{Name: "test", Limit: 4}, []string{"LoginTests/testOK", "LoginTests/testWrongPassword", "LoginTests/testRetry", "SSO/testSAML"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.TestResults(context.Background(), "my-app", "b-1", tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.want, testCasePaths(got.TestSuites))
			for _, su := range got.TestSuites {
				assert.NotEqual(t, "Empty", su.Name, "a name filter drops suites left without test cases")
			}
		})
	}

	got, err := svc.TestResults(context.Background(), "my-app", "b-1", TestResultsFilter{Name: "saml"})
	require.NoError(t, err)
	require.Len(t, got.TestSuites, 1)
	assert.Equal(t, "LoginTests", got.TestSuites[0].Name, "the parent of a matching suite stays as its container")
	assert.Empty(t, got.TestSuites[0].TestCases)
	assert.NotNil(t, got.TestSuites[0].TestCases)
	assert.Equal(t, 3, got.TestSuites[0].TestCount, "counts stay the suite's reported totals")

	got, err = svc.TestResults(context.Background(), "my-app", "b-1", TestResultsFilter{Name: "nothing-matches"})
	require.NoError(t, err)
	assert.Empty(t, got.TestSuites)
	assert.NotNil(t, got.TestSuites, "an empty result must render as [] not null")
}

func TestService_TestResults_RejectsInvalidFilters(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))

	tests := []struct {
		filter TestResultsFilter
		want   string
	}{
		{TestResultsFilter{Statuses: []string{"failure"}}, `unknown test status "failure" (expected: passed, failed, skipped, error)`},
		{TestResultsFilter{SortBy: "size"}, `unknown sort field "size" (expected: duration, name, status)`},
		{TestResultsFilter{SortBy: "name", SortOrder: "up"}, `unknown sort order "up" (expected: asc, desc)`},
		{TestResultsFilter{SortOrder: "desc"}, "a sort order needs a sort field"},
		{TestResultsFilter{Limit: -1}, "limit must be between 1 and 1000"},
		{TestResultsFilter{Limit: 1001}, "limit must be between 1 and 1000"},
	}
	for _, tt := range tests {
		_, err := svc.TestResults(context.Background(), "my-app", "b-1", tt.filter)
		require.EqualError(t, err, tt.want)
	}
}

func TestService_TestResults_MapsAPIErrors(t *testing.T) {
	tests := map[int]string{
		http.StatusBadRequest: "Couldn't fetch test results for this build. Please try again later!",
		http.StatusForbidden:  "the plan of the app's workspace does not include test reports",
		http.StatusNotFound:   `build "b-1" not found`,
		http.StatusBadGateway: "Couldn't fetch test results",
	}
	for code, want := range tests {
		srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"message":"Couldn't fetch test results for this build. Please try again later!"}`))
		})
		_, err := NewService(newAPIClient(t, srv.URL)).TestResults(context.Background(), "my-app", "b-1", TestResultsFilter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), want, "status %d", code)
	}
}

func TestService_TestResults_BadRequestReturnsServerMessage(t *testing.T) {
	for _, msg := range []string{"Test results are available once the build has finished", "Invalid limit: must be an integer between 1 and 1000"} {
		srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error_msg":"` + msg + `"}`))
		})
		_, err := NewService(newAPIClient(t, srv.URL)).TestResults(context.Background(), "my-app", "b-1", TestResultsFilter{})
		require.EqualError(t, err, msg)
	}
}

func TestService_TestResults_RequiresSlugs(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))
	_, err := svc.TestResults(context.Background(), "", "b", TestResultsFilter{})
	assert.Error(t, err)
	_, err = svc.TestResults(context.Background(), "a", "", TestResultsFilter{})
	assert.Error(t, err)
}

func TestTestResultsResult_JSONShape(t *testing.T) {
	res := TestResultsResult{TestSuites: []TestSuite{{
		Name: "S", TestCount: 1, DurationMS: 5, TestSuites: []TestSuite{},
		TestCases: []TestCase{{Name: "c", ClassName: "K", Status: "failed", Message: "m", DurationMS: 5, Flaky: true,
			Runs: []TestRun{{Name: "r", Status: "failed", Message: "m", DurationMS: 5}}}},
	}}}
	const want = `{"test_suites":[{"name":"S","test_count":1,"failure_count":0,"error_count":0,"skipped_count":0,"duration_ms":5,` +
		`"test_cases":[{"name":"c","class_name":"K","status":"failed","message":"m","duration_ms":5,"flaky":true,` +
		`"runs":[{"name":"r","status":"failed","message":"m","duration_ms":5}]}],"test_suites":[]}]}`
	b, err := json.Marshal(res)
	require.NoError(t, err)
	assert.JSONEq(t, want, string(b))

	y, err := yaml.Marshal(res)
	require.NoError(t, err)
	assert.Contains(t, string(y), "test_count: 1")
	assert.Contains(t, string(y), "duration_ms: 5")
	assert.Contains(t, string(y), "class_name: K")
	assert.Contains(t, string(y), "test_cases:")
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func newAPIClient(t *testing.T, baseURL string) *bitriseapi.Client {
	t.Helper()
	c, err := bitriseapi.New(baseURL, "t")
	require.NoError(t, err)
	return c
}

const stepsSummaryJSON = `{
  "is_log_archived": true,
  "execution": {
    "workflows": [
      {"workflow_id": "_setup", "steps": [
        {"uuid": "s-1", "step_id": "git-clone", "title": "Git clone", "status": "success",
         "started_at": "2026-10-19T13:25:35Z", "finished_at": "2026-10-19T13:25:38.793Z"}
      ]},
      {"workflow_id": "primary", "steps": [
        {"uuid": "s-2", "step_id": "script", "title": "Tests", "version": "1.2.1", "status": "failed",
         "started_at": "2026-10-19T13:26:48Z", "finished_at": "2026-10-19T13:28:33.328Z",
         "errors": [{"code": 1, "message": "exit status 1"}]},
        {"uuid": "s-3", "step_id": "save-cache", "title": "Save cache", "status": "skipped",
         "status_reason": "This Step was skipped, because a previous Step failed, and this Step was not marked \"is_always_run\"."}
      ]}
    ]
  }
}`

// newSummaryNotFoundServer answers the step summary with 404 and the build
// lookup with buildStatusCode and buildBody.
func newSummaryNotFoundServer(t *testing.T, buildStatusCode int, buildBody string) *httptest.Server {
	t.Helper()
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/my-app/builds/b-1" {
			w.WriteHeader(buildStatusCode)
			_, _ = w.Write([]byte(buildBody))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
}

// newStepLogsServer answers the step summary with stepsSummaryJSON and each step's log
// with an archive URL, whose JSON holds three messages for s-2 and none for
// any other step.
func newStepLogsServer(t *testing.T) *httptest.Server {
	t.Helper()
	archive := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/s-2" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`[{"message":"s-2 one\n","level":"normal"},{"message":"s-2 two\n","level":"normal"},{"message":"s-2 three\n","level":"normal"}]`))
	})
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if id, ok := strings.CutPrefix(r.URL.Path, "/apps/my-app/builds/b-1/log/steps/"); ok {
			if id == "s-3" {
				// build-log-service answers an empty step without an archive URL.
				_, _ = w.Write([]byte(`{"log_chunks":[]}`))
				return
			}
			_, _ = fmt.Fprintf(w, `{"expiring_raw_log_url":%q}`, archive.URL+"/"+id)
			return
		}
		body, err := filterSummaryByStatus(stepsSummaryJSON, r.URL.Query()["status"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(body)
	})
}

// filterSummaryByStatus applies the status filter of GET .../log/summary as
// build-log-service does: it keeps the steps with one of statuses, and the
// workflows left with any.
func filterSummaryByStatus(summary string, statuses []string) ([]byte, error) {
	if len(statuses) == 0 {
		return []byte(summary), nil
	}
	var s struct {
		IsLogArchived bool `json:"is_log_archived"`
		Execution     struct {
			Workflows []struct {
				WorkflowID string           `json:"workflow_id"`
				Steps      []map[string]any `json:"steps"`
			} `json:"workflows"`
		} `json:"execution"`
	}
	if err := json.Unmarshal([]byte(summary), &s); err != nil {
		return nil, err
	}
	workflows := s.Execution.Workflows[:0]
	for _, wf := range s.Execution.Workflows {
		steps := wf.Steps[:0]
		for _, st := range wf.Steps {
			if status, _ := st["status"].(string); slices.Contains(statuses, status) {
				steps = append(steps, st)
			}
		}
		if len(steps) > 0 {
			wf.Steps = steps
			workflows = append(workflows, wf)
		}
	}
	s.Execution.Workflows = workflows
	return json.Marshal(s)
}

func testCasePaths(suites []TestSuite) []string {
	var paths []string
	for _, su := range suites {
		for _, tc := range su.TestCases {
			paths = append(paths, su.Name+"/"+tc.Name)
		}
		paths = append(paths, testCasePaths(su.TestSuites)...)
	}
	return paths
}

const testResultsJSON = `{"test_suites": [
  {"name": "LoginTests", "test_count": 3, "failure_count": 1, "error_count": 0, "skipped_count": 0, "duration_in_ms": 2500,
   "properties": [{"key": "k", "value": "v"}], "attachments": [],
   "test_cases": [
     {"name": "testOK", "class_name": "LoginTests", "status": "passed", "duration_in_ms": 120, "flaky": false, "runs": []},
     {"name": "testWrongPassword", "class_name": "LoginTests", "status": "failed", "message": "expected error\nat line 12", "duration_in_ms": 1100, "flaky": false},
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
