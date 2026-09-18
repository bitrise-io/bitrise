package stack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestList_SortsByIDAndNormalizesFields(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"b-stack": {"id":"b-stack","title":"B","os":"linux","status":"stable","stack-report":"https://example.test/b","removal-date":"2027-01-01"},
			"a-stack": {"id":"a-stack","title":"A","os":"osx","status":"edge"}
		}`))
	})
	client := newAPIClient(t, srv.URL)

	result, err := NewService(client).List(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "a-stack", result.Items[0].ID)
	assert.Equal(t, "b-stack", result.Items[1].ID)
	assert.Equal(t, "https://example.test/b", result.Items[1].StackReport)
	assert.Equal(t, "2027-01-01", result.Items[1].RemovalDate)
}

func TestList_FallsBackToMapKeyWhenInfoIDEmpty(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"linux-docker-android": {"title":"Linux Android","os":"linux","status":"stable"}}`))
	})
	client := newAPIClient(t, srv.URL)

	result, err := NewService(client).List(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "linux-docker-android", result.Items[0].ID)
}

func TestList_PassesWorkspaceSlugAsOrgScopedPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	})
	client := newAPIClient(t, srv.URL)

	_, err := NewService(client).List(context.Background(), "my-workspace")
	require.NoError(t, err)
	assert.Equal(t, "/organizations/my-workspace/available-stacks", gotPath)
}

func TestStacksResult_JSONShape(t *testing.T) {
	// Field names are normalized from the wire format's hyphenated keys
	// (stack-report, removal-date) to underscored ones, and the whole list
	// is wrapped in {"items": [...]}, matching the shape bitrise-cli's
	// `stack list --output json` already ships to beta users.
	result := StacksResult{Items: []Stack{{
		ID: "linux-docker-android", Title: "Linux Android", OS: "linux",
		Status: "stable", StackReport: "https://example.test/report", RemovalDate: "2027-01-01",
	}}}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))
	items, ok := got["items"].([]any)
	require.True(t, ok, "expected top-level \"items\" array, got %s", data)
	require.Len(t, items, 1)
	item := items[0].(map[string]any)
	assert.Equal(t, "https://example.test/report", item["stack_report"])
	assert.Equal(t, "2027-01-01", item["removal_date"])
	assert.NotContains(t, string(data), "stack-report")
	assert.NotContains(t, string(data), "removal-date")
}

func TestList_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})
	client := newAPIClient(t, srv.URL)

	_, err := NewService(client).List(context.Background(), "")
	require.Error(t, err)
	apiErr, ok := err.(*bitriseapi.APIError)
	require.True(t, ok, "expected *bitriseapi.APIError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
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
