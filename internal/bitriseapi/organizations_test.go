package bitriseapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizations_ParsesResponse(t *testing.T) {
	var gotPath, gotAuth string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[{"slug":"acme","name":"Acme"}]}`))
	})

	orgs, err := newAPIClient(t, srv.URL, "my-token").Organizations(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/organizations", gotPath)
	assert.Equal(t, "token my-token", gotAuth)
	assert.Equal(t, []Organization{{Slug: "acme", Name: "Acme"}}, orgs)
}

func TestOrganizations_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").Organizations(context.Background())
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}

func TestOrganization_ParsesResponse(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"data":{"slug":"acme/x","name":"Acme"}}`))
	})

	org, err := newAPIClient(t, srv.URL, "t").Organization(context.Background(), "acme/x")
	require.NoError(t, err)
	assert.Equal(t, "/organizations/acme%2Fx", gotPath)
	assert.Equal(t, Organization{Slug: "acme/x", Name: "Acme"}, org)
}

func TestOrganization_PropagatesNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").Organization(context.Background(), "missing")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
