package bitriseapi

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApps_ParsesResponseAndPagingCursor(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"data": [{"slug":"app-1","title":"First","provider":"github","repo_url":"https://github.com/x/y","project_type":"android","is_disabled":false,"owner":{"slug":"acme"}}],
			"paging": {"next":"page-2"}
		}`))
	})

	apps, next, err := newAPIClient(t, srv.URL, "t").Apps(context.Background(), AppsListOptions{})
	require.NoError(t, err)
	require.Len(t, apps, 1)
	assert.Equal(t, App{Slug: "app-1", Title: "First", Provider: "github", RepoURL: "https://github.com/x/y", ProjectType: "android", Owner: AppOwner{Slug: "acme"}}, apps[0])
	assert.Equal(t, "page-2", next)
}

func TestApps_SendsListOptionsAsQueryParams(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	_, _, err := newAPIClient(t, srv.URL, "t").Apps(context.Background(), AppsListOptions{
		SortBy:      "created_at",
		Next:        "cur",
		Limit:       10,
		Title:       "exact-title",
		ProjectType: "android",
	})
	require.NoError(t, err)
	q, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "created_at", q.Get("sort_by"))
	assert.Equal(t, "cur", q.Get("next"))
	assert.Equal(t, "10", q.Get("limit"))
	assert.Equal(t, "exact-title", q.Get("title"))
	assert.Equal(t, "android", q.Get("project_type"))
}

func TestApps_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})

	_, _, err := newAPIClient(t, srv.URL, "t").Apps(context.Background(), AppsListOptions{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestApp_HitsCorrectPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"my-app","title":"App","provider":"github","owner":{"slug":"acme"}}}`))
	})

	app, err := newAPIClient(t, srv.URL, "t").App(context.Background(), "my-app")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app", gotPath)
	assert.Equal(t, App{Slug: "my-app", Title: "App", Provider: "github", Owner: AppOwner{Slug: "acme"}}, app)
}

func TestApp_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").App(context.Background(), "missing")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestRegisterApp(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"slug":"new-app"}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").RegisterApp(context.Background(), RegisterAppRequest{
		RepoURL:           "https://github.com/acme/widget.git",
		OrganizationSlug:  "acme",
		Provider:          "custom",
		DefaultBranchName: "main",
		FlowType:          "cli",
	})
	require.NoError(t, err)
	assert.Equal(t, "/apps/register", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, RegisterAppResponse{Slug: "new-app"}, resp)
	assert.JSONEq(t, `{"repo_url":"https://github.com/acme/widget.git","organization_slug":"acme","provider":"custom","is_public":false,"default_branch_name":"main","flow_type":"cli"}`, gotBody)
}

func TestFinishApp(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"build_trigger_token":"btt","branch_name":"main"}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").FinishApp(context.Background(), "new-app", FinishAppRequest{StackID: "linux-docker-android-22.04", Mode: "manual"})
	require.NoError(t, err)
	assert.Equal(t, "/apps/new-app/finish", gotPath)
	assert.Equal(t, FinishAppResponse{BuildTriggerToken: "btt", BranchName: "main"}, resp)
}
