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

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "")
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

func TestService_Steps_FiltersByStatus(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "failed")
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2", got.Items[0].ID)

	got, err = svc.Steps(context.Background(), "my-app", "b-1", "preparation_failed")
	require.NoError(t, err)
	assert.Empty(t, got.Items)
	assert.NotNil(t, got.Items, "an empty result must render as [] not null")
}

func TestService_Steps_FailedMatchesEveryWayOfFailingTheBuild(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_log_archived": true, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "ok", "status": "success"},
  {"uuid": "failed", "status": "failed"},
  {"uuid": "prep", "status": "preparation_failed"},
  {"uuid": "timeout", "status": "aborted_with_custom_timeout"},
  {"uuid": "no-output", "status": "aborted_with_no_output"},
  {"uuid": "skippable", "status": "failed_skippable"},
  {"uuid": "aborted", "status": "aborted", "started_at": "2026-10-19T14:15:26Z", "finished_at": "2026-10-19T14:16:56.687Z"},
  {"uuid": "after-abort", "status": "not_started", "started_at": null, "finished_at": null}
]}]}}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	for status, want := range map[string][]string{
		"failed":             {"failed", "prep", "timeout", "no-output"},
		"preparation_failed": {"prep"},
		"failed_skippable":   {"skippable"},
		"aborted":            {"aborted"},
		"not_started":        {"after-abort"},
	} {
		got, err := svc.Steps(context.Background(), "my-app", "b-1", status)
		require.NoError(t, err)
		var ids []string
		for _, st := range got.Items {
			ids = append(ids, st.ID)
		}
		assert.Equal(t, want, ids, "status %q", status)
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

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "")
	require.NoError(t, err)
	assert.False(t, got.Complete)
	require.Len(t, got.Items, 2)
	assert.Equal(t, "started", got.Items[1].Status)
	assert.Nil(t, got.Items[1].FinishedAt)
}

func TestService_Steps_RequiresSlugs(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))

	_, err := svc.Steps(context.Background(), "", "b-1", "")
	require.EqualError(t, err, "app ID is required")
	_, err = svc.Steps(context.Background(), "my-app", "", "")
	require.EqualError(t, err, "build ID is required")
}

func TestService_Steps_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "missing", "")
	require.EqualError(t, err, `build "missing" not found`)
}

func TestService_Steps_NoSummaryYetForRunningBuild(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusOK, `{"data":{"slug":"b-1","status":0}}`)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.Steps(context.Background(), "my-app", "b-1", "")
	require.NoError(t, err)
	assert.False(t, got.Complete)
	assert.NotNil(t, got.Items)
	assert.Empty(t, got.Items)
}

func TestService_Steps_NoSummaryForFinishedBuild(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusOK, `{"data":{"slug":"b-1","status":3}}`)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "")
	require.EqualError(t, err, `build "b-1" has no step summary`)
}

func TestService_Steps_NoSummaryAndBuildLookupFails(t *testing.T) {
	srv := newSummaryNotFoundServer(t, http.StatusInternalServerError, `{"message":"boom"}`)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "")
	require.ErrorContains(t, err, `no step summary for build "b-1", and looking up the build failed`)
}

func TestService_Steps_RejectsUnknownStatus(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))

	_, err := svc.Steps(context.Background(), "my-app", "b-1", "aborted-with-success")
	require.ErrorContains(t, err, `unknown step status "aborted-with-success" (expected: not_started, started, success`)
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

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "failed", nil, 0)
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

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", []string{"s-3", "s-1"}, 0)
	require.NoError(t, err)
	require.Len(t, got.Items, 2)
	assert.Equal(t, "s-1", got.Items[0].ID, "steps keep execution order, not the order of the IDs")
	assert.Equal(t, "s-3", got.Items[1].ID)
	assert.Empty(t, got.Items[1].Log, "a step that logged nothing has an empty log")

	got, err = svc.StepLogs(context.Background(), "my-app", "b-1", "failed", []string{"s-1"}, 0)
	require.NoError(t, err)
	assert.Empty(t, got.Items)
	assert.NotNil(t, got.Items, "an empty result must render as [] not null")
}

func TestService_StepLogs_Tail(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", []string{"s-2"}, 2)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2 two\ns-2 three\n", got.Items[0].Log)

	_, err = svc.StepLogs(context.Background(), "my-app", "b-1", "", nil, -1)
	require.Error(t, err)
}

func TestService_StepLogs_UnknownStepID(t *testing.T) {
	srv := newStepLogsServer(t)
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", []string{"s-2", "nope"}, 0)
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

	got, err := svc.StepLogs(context.Background(), "my-app", "b-1", "", []string{"s-1", "not-started-yet"}, 0)
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
		_, _ = w.Write([]byte(stepsSummaryJSON))
	})
}
