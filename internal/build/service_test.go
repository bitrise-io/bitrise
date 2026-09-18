package build

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
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

	_, err := svc.Trigger(context.Background(), TriggerRequest{
		AppSlug:       "my-app",
		Workflow:      "primary",
		PullRequestID: 42,
		Priority:      1,
		Environments: []TriggerEnv{
			{Key: "MY_VAR", Value: "hello"},
		},
	})
	require.NoError(t, err)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, float64(42), bp["pull_request_id"])
	assert.Equal(t, float64(1), bp["priority"])
	envs, _ := bp["environments"].([]any)
	require.Len(t, envs, 1)
	env, _ := envs[0].(map[string]any)
	assert.Equal(t, "MY_VAR", env["mapped_to"])
	assert.Equal(t, "hello", env["value"])
	assert.Equal(t, true, env["is_expand"])
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
	require.NoError(t, svc.Log(context.Background(), "my-app", "b-1", &buf))
	assert.Contains(t, buf.String(), "LOG OUTPUT")
}

func TestService_Log_RequiresSlugs(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})
	svc := NewService(newAPIClient(t, srv.URL))

	assert.Error(t, svc.Log(context.Background(), "", "b", io.Discard))
	assert.Error(t, svc.Log(context.Background(), "a", "", io.Discard))
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

func TestService_NilClientFails(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.List(context.Background(), ListOptions{AppSlug: "x"})
	assert.Error(t, err)
	_, err = svc.View(context.Background(), "a", "b")
	assert.Error(t, err)
	_, err = svc.Trigger(context.Background(), TriggerRequest{AppSlug: "a"})
	assert.Error(t, err)
	assert.Error(t, svc.Log(context.Background(), "a", "b", io.Discard))
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
