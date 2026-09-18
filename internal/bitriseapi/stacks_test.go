package bitriseapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableStacks_GlobalPath(t *testing.T) {
	var gotPath, gotAuth string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := newAPIClient(t, srv.URL, "my-token").AvailableStacks(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "/available-stacks", gotPath)
	assert.Equal(t, "token my-token", gotAuth)
}

func TestAvailableStacks_OrgScopedPath(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").AvailableStacks(context.Background(), "my-workspace")
	require.NoError(t, err)
	assert.Equal(t, "/organizations/my-workspace/available-stacks", gotPath)
}

func TestAvailableStacks_EscapesOrgSlug(t *testing.T) {
	// The slug reaches the path unvalidated, so it must be escaped: an
	// unescaped "/" would silently split into extra path segments, and ".."
	// would let a slug walk out of /organizations and hit another endpoint.
	tests := map[string]string{
		"a/b":         "/organizations/a%2Fb/available-stacks",
		"../../admin": "/organizations/..%2F..%2Fadmin/available-stacks",
		"a b":         "/organizations/a%20b/available-stacks",
	}
	for slug, wantURI := range tests {
		t.Run(slug, func(t *testing.T) {
			var gotURI string
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotURI = r.RequestURI
				_, _ = w.Write([]byte(`{}`))
			})

			_, err := newAPIClient(t, srv.URL, "t").AvailableStacks(context.Background(), slug)
			require.NoError(t, err)
			assert.Equal(t, wantURI, gotURI)
		})
	}
}

func TestAvailableStacks_ParsesResponse(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"linux-docker-android": {"id":"linux-docker-android","title":"Linux Android","os":"linux","status":"stable","stack-report":"https://example.test/report","removal-date":"2027-01-01"}
		}`))
	})

	stacks, err := newAPIClient(t, srv.URL, "t").AvailableStacks(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, stacks, 1)
	got := stacks["linux-docker-android"]
	assert.Equal(t, "linux-docker-android", got.ID)
	assert.Equal(t, "Linux Android", got.Title)
	assert.Equal(t, "stable", got.Status)
	assert.Equal(t, "https://example.test/report", got.StackReport)
	assert.Equal(t, "2027-01-01", got.RemovalDate)
}

func TestAvailableStacks_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").AvailableStacks(context.Background(), "")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Equal(t, "forbidden", apiErr.Message)
}
