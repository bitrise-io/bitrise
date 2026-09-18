package bitriseapi

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppBitriseYML(t *testing.T) {
	var gotPath, gotAuth, gotAccept string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	})

	content, err := newAPIClient(t, srv.URL, "my-token").AppBitriseYML(context.Background(), "app-slug")
	require.NoError(t, err)
	assert.Equal(t, "/apps/app-slug/bitrise.yml", gotPath)
	assert.Equal(t, "token my-token", gotAuth)
	assert.Equal(t, "text/plain", gotAccept)
	assert.Equal(t, "format_version: \"13\"\n", content)
}

func TestBuildBitriseYML(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	})

	_, err := newAPIClient(t, srv.URL, "t").BuildBitriseYML(context.Background(), "app-slug", "build-slug")
	require.NoError(t, err)
	assert.Equal(t, "/apps/app-slug/builds/build-slug/bitrise.yml", gotPath)
}

func TestUpdateAppBitriseYML(t *testing.T) {
	var gotPath, gotMethod, gotContentType, gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{}`))
	})

	err := newAPIClient(t, srv.URL, "t").UpdateAppBitriseYML(context.Background(), "app-slug", map[string]any{"format_version": "13"})
	require.NoError(t, err)
	assert.Equal(t, "/apps/app-slug/bitrise.yml", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json", gotContentType)
	assert.JSONEq(t, `{"app_config_datastore_yaml":{"format_version":"13"}}`, gotBody)
}

func TestUpdateAppBitriseYML_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"app not found"}`))
	})

	err := newAPIClient(t, srv.URL, "t").UpdateAppBitriseYML(context.Background(), "app-slug", map[string]any{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestValidateBitriseYML_NoAppSlug(t *testing.T) {
	var gotQuery string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"errors":[],"warnings":[]}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").ValidateBitriseYML(context.Background(), "format_version: \"13\"\n", "")
	require.NoError(t, err)
	assert.Empty(t, gotQuery)
	assert.Empty(t, resp.Errors)
	assert.Empty(t, resp.Warnings)
}

func TestValidateBitriseYML_WithAppSlug(t *testing.T) {
	var gotQuery, gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"errors":["missing format_version"],"warnings":["deprecated step used"]}`))
	})

	resp, err := newAPIClient(t, srv.URL, "t").ValidateBitriseYML(context.Background(), "workflows: {}\n", "app-slug")
	require.NoError(t, err)
	assert.Equal(t, "app_slug=app-slug", gotQuery)
	assert.JSONEq(t, `{"bitrise_yml":"workflows: {}\n"}`, gotBody)
	assert.Equal(t, []string{"missing format_version"}, resp.Errors)
	assert.Equal(t, []string{"deprecated step used"}, resp.Warnings)
}

func TestValidateBitriseYML_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"could not parse YAML"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").ValidateBitriseYML(context.Background(), "not: valid: yaml:", "")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.StatusCode)
	assert.Equal(t, "could not parse YAML", apiErr.Message)
}
