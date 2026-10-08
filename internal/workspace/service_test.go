package workspace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
)

func TestServiceList_SortsAndMapsSlugToID(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"b-ws","name":"Bravo"},{"slug":"a-ws","name":"Alpha"}]}`))
	})

	result, err := NewService(newAPIClient(t, srv.URL)).List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []Workspace{{ID: "a-ws", Name: "Alpha"}, {ID: "b-ws", Name: "Bravo"}}, result.Items)
}

func TestServiceList_EmptyIsNotAnError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	result, err := NewService(newAPIClient(t, srv.URL)).List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Items)
}

func TestServiceList_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})

	_, err := NewService(newAPIClient(t, srv.URL)).List(context.Background())
	var apiErr *bitriseapi.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestServiceView(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"a-ws","name":"Alpha"}}`))
	})

	ws, err := NewService(newAPIClient(t, srv.URL)).View(context.Background(), "a-ws")
	require.NoError(t, err)
	assert.Equal(t, "/organizations/a-ws", gotPath)
	assert.Equal(t, Workspace{ID: "a-ws", Name: "Alpha"}, ws)
}

func TestServiceView_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	})

	_, err := NewService(newAPIClient(t, srv.URL)).View(context.Background(), "missing")
	require.EqualError(t, err, `workspace "missing" not found`)
}

func TestWorkspacesResult_Shape(t *testing.T) {
	result := WorkspacesResult{Items: []Workspace{{ID: "a-ws", Name: "Alpha"}}}
	want := map[string]any{"items": []any{map[string]any{"id": "a-ws", "name": "Alpha"}}}

	jsonData, err := json.Marshal(result)
	require.NoError(t, err)
	var gotJSON map[string]any
	require.NoError(t, json.Unmarshal(jsonData, &gotJSON))
	assert.Equal(t, want, gotJSON)

	yamlData, err := yaml.Marshal(result)
	require.NoError(t, err)
	assert.Equal(t, "items:\n- id: a-ws\n  name: Alpha\n", string(yamlData))
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
