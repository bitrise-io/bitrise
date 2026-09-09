package resolve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/cache"
)

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func newAPIClient(t *testing.T, baseURL string) *bitriseapi.Client {
	t.Helper()
	c, err := bitriseapi.New(baseURL, "test-token")
	require.NoError(t, err)
	return c
}

func appsBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps" {
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}
}

func TestAppSlug_LiteralSlugPassthrough(t *testing.T) {
	var called bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		_, _ = w.Write([]byte(`{"data":[],"paging":{}}`))
	})
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.AppSlug(context.Background(), "abc12345")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
	assert.True(t, called, "expected API to be called even for slug-like input")
}

func TestAppSlug_NameResolution(t *testing.T) {
	var gotTitle string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotTitle, _ = url.QueryUnescape(r.URL.Query().Get("title"))
		_, _ = w.Write([]byte(`{"data":[{"slug":"abc12345","title":"My App","owner":{}}],"paging":{}}`))
	})
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.AppSlug(context.Background(), "My App")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
	assert.Equal(t, "My App", gotTitle)
}

func TestAppSlug_CaseInsensitive(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[{"slug":"abc12345","title":"My App","owner":{}}],"paging":{}}`))
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.AppSlug(context.Background(), "my app")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
}

func TestAppSlug_AmbiguousError(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[
		{"slug":"slug-1","title":"My App","owner":{}},
		{"slug":"slug-2","title":"My App","owner":{}}
	],"paging":{}}`))
	r := New(newAPIClient(t, srv.URL), nil)

	_, err := r.AppSlug(context.Background(), "My App")
	require.Error(t, err)
}

func TestAppSlug_CacheHit(t *testing.T) {
	var apiCalled bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		_, _ = w.Write([]byte(`{"data":[],"paging":{}}`))
	})
	c := cache.New()
	c.SetApp("My App", "cached-slug")
	r := New(newAPIClient(t, srv.URL), c)

	slug, err := r.AppSlug(context.Background(), "My App")
	require.NoError(t, err)
	assert.Equal(t, "cached-slug", slug)
	assert.False(t, apiCalled, "API should not be called on cache hit")
}

func TestAppSlug_PopulatesCache(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[{"slug":"abc12345","title":"My App","owner":{}}],"paging":{}}`))
	c := cache.New()
	r := New(newAPIClient(t, srv.URL), c)

	_, err := r.AppSlug(context.Background(), "My App")
	require.NoError(t, err)

	slug, ok := c.LookupApp("My App")
	assert.True(t, ok)
	assert.Equal(t, "abc12345", slug)
}

func TestResolveApp_NameMatch_ReturnsFull(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[{"slug":"abc12345","title":"My App","provider":"github","repo_url":"https://github.com/x/y","owner":{"slug":"acme"}}],"paging":{}}`))
	r := New(newAPIClient(t, srv.URL), nil)

	app, complete, err := r.ResolveApp(context.Background(), "My App")
	require.NoError(t, err)
	assert.True(t, complete, "expected complete=true on name match")
	assert.Equal(t, "abc12345", app.Slug)
	assert.Equal(t, "My App", app.Title)
	assert.Equal(t, "github", app.Provider)
}

func TestResolveApp_NoMatch_Passthrough(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[],"paging":{}}`))
	r := New(newAPIClient(t, srv.URL), nil)

	app, complete, err := r.ResolveApp(context.Background(), "abc12345")
	require.NoError(t, err)
	assert.False(t, complete, "expected complete=false on passthrough")
	assert.Equal(t, "abc12345", app.Slug)
}

func TestResolveApp_CacheHit_NotFetched(t *testing.T) {
	var apiCalled bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		_, _ = w.Write([]byte(`{"data":[],"paging":{}}`))
	})
	c := cache.New()
	c.SetApp("My App", "cached-slug")
	r := New(newAPIClient(t, srv.URL), c)

	app, complete, err := r.ResolveApp(context.Background(), "My App")
	require.NoError(t, err)
	assert.False(t, complete, "expected complete=false on cache hit (full data not cached)")
	assert.Equal(t, "cached-slug", app.Slug)
	assert.False(t, apiCalled, "API should not be called on cache hit")
}

func TestResolveApp_AmbiguousError(t *testing.T) {
	srv := newFakeServer(t, appsBody(`{"data":[
		{"slug":"slug-1","title":"My App","owner":{}},
		{"slug":"slug-2","title":"My App","owner":{}}
	],"paging":{}}`))
	r := New(newAPIClient(t, srv.URL), nil)

	_, _, err := r.ResolveApp(context.Background(), "My App")
	require.Error(t, err)
}

func organizationsBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/organizations" {
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}
}

func TestWorkspaceSlug_LiteralSlugPassthrough(t *testing.T) {
	var called bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.WorkspaceSlug(context.Background(), "abc12345")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
	assert.True(t, called, "expected API to be called even for slug-like input")
}

func TestWorkspaceSlug_NameResolution(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[{"slug":"abc12345","name":"My Workspace"}]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.WorkspaceSlug(context.Background(), "My Workspace")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
}

func TestWorkspaceSlug_CaseInsensitive(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[{"slug":"abc12345","name":"My Workspace"}]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	slug, err := r.WorkspaceSlug(context.Background(), "my workspace")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", slug)
}

func TestWorkspaceSlug_AmbiguousError(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[
		{"slug":"slug-1","name":"My Workspace"},
		{"slug":"slug-2","name":"My Workspace"}
	]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	_, err := r.WorkspaceSlug(context.Background(), "My Workspace")
	require.Error(t, err)
}

func TestWorkspaceSlug_CacheHit(t *testing.T) {
	var apiCalled bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	c := cache.New()
	c.SetWorkspace("My Workspace", "cached-slug")
	r := New(newAPIClient(t, srv.URL), c)

	slug, err := r.WorkspaceSlug(context.Background(), "My Workspace")
	require.NoError(t, err)
	assert.Equal(t, "cached-slug", slug)
	assert.False(t, apiCalled, "API should not be called on cache hit")
}

func TestWorkspaceSlug_PopulatesCache(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[{"slug":"abc12345","name":"My Workspace"}]}`))
	c := cache.New()
	r := New(newAPIClient(t, srv.URL), c)

	_, err := r.WorkspaceSlug(context.Background(), "My Workspace")
	require.NoError(t, err)

	slug, ok := c.LookupWorkspace("My Workspace")
	assert.True(t, ok)
	assert.Equal(t, "abc12345", slug)
}

func TestResolveWorkspace_NameMatch_ReturnsFull(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[{"slug":"abc12345","name":"My Workspace"}]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	org, complete, err := r.ResolveWorkspace(context.Background(), "My Workspace")
	require.NoError(t, err)
	assert.True(t, complete, "expected complete=true on name match")
	assert.Equal(t, "abc12345", org.Slug)
	assert.Equal(t, "My Workspace", org.Name)
}

func TestResolveWorkspace_NoMatch_Passthrough(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	org, complete, err := r.ResolveWorkspace(context.Background(), "abc12345")
	require.NoError(t, err)
	assert.False(t, complete, "expected complete=false on passthrough")
	assert.Equal(t, "abc12345", org.Slug)
}

func TestResolveWorkspace_CacheHit_NotFetched(t *testing.T) {
	var apiCalled bool
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	c := cache.New()
	c.SetWorkspace("My Workspace", "cached-slug")
	r := New(newAPIClient(t, srv.URL), c)

	org, complete, err := r.ResolveWorkspace(context.Background(), "My Workspace")
	require.NoError(t, err)
	assert.False(t, complete, "expected complete=false on cache hit (full data not cached)")
	assert.Equal(t, "cached-slug", org.Slug)
	assert.False(t, apiCalled, "API should not be called on cache hit")
}

func TestResolveWorkspace_AmbiguousError(t *testing.T) {
	srv := newFakeServer(t, organizationsBody(`{"data":[
		{"slug":"slug-1","name":"My Workspace"},
		{"slug":"slug-2","name":"My Workspace"}
	]}`))
	r := New(newAPIClient(t, srv.URL), nil)

	_, _, err := r.ResolveWorkspace(context.Background(), "My Workspace")
	require.Error(t, err)
}
