package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestList_MapsAPIShape(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"data": [
				{"slug":"app-1","title":"First","provider":"github","repo_url":"https://github.com/x/y","project_type":"android","is_disabled":false,"owner":{"slug":"acme"}},
				{"slug":"app-2","title":"Second","provider":"gitlab","repo_url":"https://gitlab.com/x/y","project_type":"ios","is_disabled":true,"owner":{"slug":"bob"}}
			],
			"paging": {"next":"page-2"}
		}`))
	})
	client := newAPIClient(t, srv.URL)

	res, err := NewService(client).List(context.Background(), ListOptions{})
	require.NoError(t, err)
	require.Len(t, res.Items, 2)
	assert.Equal(t, App{Slug: "app-1", Title: "First", Provider: "github", RepoURL: "https://github.com/x/y", OwnerSlug: "acme", ProjectType: "android"}, res.Items[0])
	assert.True(t, res.Items[1].IsDisabled)
	assert.Equal(t, "page-2", res.NextCursor)
}

func TestList_PassesOptionsAsQueryParams(t *testing.T) {
	var gotQuery url.Values
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	client := newAPIClient(t, srv.URL)

	_, err := NewService(client).List(context.Background(), ListOptions{
		Limit: 10, Cursor: "cur", SortBy: "created_at", Title: "exact-title", ProjectType: "android",
	})
	require.NoError(t, err)
	assert.Equal(t, "10", gotQuery.Get("limit"))
	assert.Equal(t, "cur", gotQuery.Get("next"))
	assert.Equal(t, "created_at", gotQuery.Get("sort_by"))
	assert.Equal(t, "exact-title", gotQuery.Get("title"))
	assert.Equal(t, "android", gotQuery.Get("project_type"))
}

func TestList_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})
	client := newAPIClient(t, srv.URL)

	_, err := NewService(client).List(context.Background(), ListOptions{})
	require.Error(t, err)
	apiErr, ok := err.(*bitriseapi.APIError)
	require.True(t, ok, "expected *bitriseapi.APIError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestView_HitsCorrectPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"my-app","title":"App","provider":"github","owner":{"slug":"acme"}}}`))
	})
	client := newAPIClient(t, srv.URL)

	got, err := NewService(client).View(context.Background(), "my-app")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app", gotPath)
	assert.Equal(t, App{Slug: "my-app", Title: "App", Provider: "github", OwnerSlug: "acme"}, got)
}

func TestView_AppNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	client := newAPIClient(t, srv.URL)

	_, err := NewService(client).View(context.Background(), "missing-app")
	require.EqualError(t, err, `app "missing-app" not found`)
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
