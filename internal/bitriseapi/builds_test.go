package bitriseapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilds_PathAndQueryParams(t *testing.T) {
	var gotPath, gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[],"paging":{"next":""}}`))
	})

	status := 1
	pipelineBuild := true
	_, _, err := newAPIClient(t, srv.URL, "t").Builds(context.Background(), "my-app", BuildsListOptions{
		Branch:          "main",
		Workflow:        "deploy",
		Status:          &status,
		IsPipelineBuild: &pipelineBuild,
		Limit:           20,
		Next:            "cur",
	})
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds", gotPath)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "main", q.Get("branch"))
	assert.Equal(t, "deploy", q.Get("workflow"))
	assert.Equal(t, "1", q.Get("status"))
	assert.Equal(t, "true", q.Get("is_pipeline_build"))
	assert.Equal(t, "20", q.Get("limit"))
	assert.Equal(t, "cur", q.Get("next"))
}

func TestBuilds_StatusZeroIsValidFilter(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	zero := 0
	_, _, err := newAPIClient(t, srv.URL, "t").Builds(context.Background(), "my-app", BuildsListOptions{Status: &zero})
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "0", q.Get("status"))
}

func TestBuilds_NoStatusOmitsParam(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	_, _, err := newAPIClient(t, srv.URL, "t").Builds(context.Background(), "my-app", BuildsListOptions{})
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.False(t, q.Has("status"), "status query param should be omitted when filter is nil")
}

func TestBuilds_ParsesResponse(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "data": [
    {
      "slug": "build-1",
      "build_number": 42,
      "status": 1,
      "status_text": "success",
      "branch": "main",
      "triggered_workflow": "primary",
      "commit_hash": "deadbeef",
      "triggered_at": "2026-05-06T10:00:00Z",
      "finished_at": "2026-05-06T10:05:00Z"
    }
  ],
  "paging": {"next": "cursor-abc"}
}`))
	})

	builds, next, err := newAPIClient(t, srv.URL, "t").Builds(context.Background(), "my-app", BuildsListOptions{})
	require.NoError(t, err)
	require.Len(t, builds, 1)
	b := builds[0]
	assert.Equal(t, "build-1", b.Slug)
	assert.Equal(t, 42, b.BuildNumber)
	assert.Equal(t, 1, b.Status)
	assert.False(t, b.TriggeredAt.IsZero())
	assert.False(t, b.FinishedAt.IsZero())
	assert.Equal(t, "cursor-abc", next)
}

func TestBuild_Single(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"build-xyz","build_number":7,"status":2,"status_text":"failed","triggered_workflow":"deploy"}}`))
	})

	b, err := newAPIClient(t, srv.URL, "t").Build(context.Background(), "my-app", "build-xyz")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/builds/build-xyz", gotPath)
	assert.Equal(t, "build-xyz", b.Slug)
	assert.Equal(t, 7, b.BuildNumber)
	assert.Equal(t, 2, b.Status)
	assert.Equal(t, "deploy", b.TriggeredWorkflow)
}

func TestTriggerBuild_RequestAndResponse(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
  "build_slug": "new-build-001",
  "build_number": 99,
  "build_url": "https://app.bitrise.io/build/new-build-001",
  "triggered_workflow": "deploy",
  "status": "ok",
  "message": "Build triggered"
}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").TriggerBuild(context.Background(), "my-app", TriggerBuildRequest{
		HookInfo: TriggerBuildHookInfo{Type: "bitrise"},
		BuildParams: TriggerBuildParams{
			WorkflowID: "deploy",
			Branch:     "release/2.0",
			CommitHash: "abc123",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds", gotPath)
	assert.Equal(t, "application/json", gotContentType)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	hi, _ := sent["hook_info"].(map[string]any)
	assert.Equal(t, "bitrise", hi["type"])
	bp, _ := sent["build_params"].(map[string]any)
	assert.Equal(t, "deploy", bp["workflow_id"])
	assert.Equal(t, "release/2.0", bp["branch"])
	assert.Equal(t, "abc123", bp["commit_hash"])

	assert.Equal(t, "new-build-001", resp.BuildSlug)
	assert.Equal(t, 99, resp.BuildNumber)
	assert.NotEmpty(t, resp.BuildURL)
}

func TestTriggerBuild_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"workflow_id is required"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").TriggerBuild(context.Background(), "my-app", TriggerBuildRequest{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestAbortBuild_RequestAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").AbortBuild(context.Background(), "my-app", "build-1", AbortBuildRequest{
		AbortReason:      "no longer needed",
		AbortWithSuccess: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds/build-1/abort", gotPath)
	assert.Equal(t, AbortBuildResponse{Status: "ok"}, resp)
	assert.JSONEq(t, `{"abort_reason":"no longer needed","abort_with_success":true}`, string(gotBody))
}

func TestAbortBuild_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"build not found"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").AbortBuild(context.Background(), "my-app", "missing", AbortBuildRequest{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestBuildLog_StreamsArchivedURL(t *testing.T) {
	var gotAuthHeader string
	rawSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("FULL ARCHIVED LOG CONTENT\n"))
	}))
	t.Cleanup(rawSrv.Close)

	apiSrv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_archived":          true,
			"expiring_raw_log_url": rawSrv.URL,
		})
	})

	var buf bytes.Buffer
	manifest, err := newAPIClient(t, apiSrv.URL, "t").BuildLog(context.Background(), "my-app", "my-build", &buf)
	require.NoError(t, err)
	assert.True(t, manifest.IsArchived)
	assert.Contains(t, buf.String(), "FULL ARCHIVED LOG CONTENT")
	assert.Empty(t, gotAuthHeader, "the presigned URL must not receive our Authorization header")
}

func TestBuildLog_StreamsChunksWhenNotArchived(t *testing.T) {
	apiSrv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "is_archived": false,
  "log_chunks": [
    {"chunk": "first chunk\n", "position": 0},
    {"chunk": "second chunk\n", "position": 1}
  ]
}`))
	})

	var buf bytes.Buffer
	manifest, err := newAPIClient(t, apiSrv.URL, "t").BuildLog(context.Background(), "my-app", "my-build", &buf)
	require.NoError(t, err)
	assert.False(t, manifest.IsArchived)
	assert.Contains(t, buf.String(), "first chunk")
	assert.Contains(t, buf.String(), "second chunk")
}

func TestBuildLog_OutOfOrderChunksAreSorted(t *testing.T) {
	apiSrv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "is_archived": false,
  "log_chunks": [
    {"chunk": "b\n", "position": 1},
    {"chunk": "a\n", "position": 0}
  ]
}`))
	})

	var buf bytes.Buffer
	_, err := newAPIClient(t, apiSrv.URL, "t").BuildLog(context.Background(), "my-app", "my-build", &buf)
	require.NoError(t, err)
	assert.Equal(t, "a\nb\n", buf.String())
}

func TestBuildLog_PropagatesManifestError(t *testing.T) {
	apiSrv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"build not found"}`))
	})

	var buf bytes.Buffer
	_, err := newAPIClient(t, apiSrv.URL, "t").BuildLog(context.Background(), "my-app", "missing", &buf)
	require.Error(t, err)
	assert.Empty(t, buf.String(), "buffer should be untouched on error")
}

func TestBuildLogManifest_SendsAfterTimestampParam(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"is_archived":false}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").BuildLogManifest(context.Background(), "my-app", "my-build", "ts-abc")
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "ts-abc", q.Get("after_timestamp"))
}

func TestBuildLogManifest_OmitsAfterTimestampWhenEmpty(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"is_archived":false}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").BuildLogManifest(context.Background(), "my-app", "my-build", "")
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.False(t, q.Has("after_timestamp"), "after_timestamp should be omitted on first call")
}

func TestBuildLogManifest_ParsesNextAfterTimestamp(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
  "is_archived": false,
  "next_after_timestamp": "2026-05-06T10:01:00Z",
  "log_chunks": [{"chunk": "hello\n", "position": 0}]
}`))
	})

	manifest, err := newAPIClient(t, srv.URL, "t").BuildLogManifest(context.Background(), "my-app", "my-build", "")
	require.NoError(t, err)
	assert.Equal(t, "2026-05-06T10:01:00Z", manifest.NextAfterTimestamp)
	require.Len(t, manifest.LogChunks, 1)
	assert.Equal(t, "hello\n", manifest.LogChunks[0].Chunk)
}
