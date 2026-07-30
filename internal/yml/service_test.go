package yml

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestGet_App(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/apps/app-slug/bitrise.yml", r.URL.Path)
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Get(context.Background(), "app-slug", "")
	require.NoError(t, err)
	assert.Equal(t, "app-slug", result.AppSlug)
	assert.Empty(t, result.BuildSlug)
	assert.Equal(t, "format_version: \"13\"\n", result.Content)
}

func TestGet_Build(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/apps/app-slug/builds/build-slug/bitrise.yml", r.URL.Path)
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Get(context.Background(), "app-slug", "build-slug")
	require.NoError(t, err)
	assert.Equal(t, "build-slug", result.BuildSlug)
}

func TestGet_AppNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	_, err := NewService(bitriseapi.New(srv.URL, "t")).Get(context.Background(), "app-slug", "")
	require.EqualError(t, err, `app "app-slug" not found`)
}

func TestGet_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	_, err := NewService(bitriseapi.New(srv.URL, "t")).Get(context.Background(), "app-slug", "build-slug")
	require.EqualError(t, err, `build "build-slug" not found`)
}

func TestGet_PropagatesOtherErrors(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	})

	_, err := NewService(bitriseapi.New(srv.URL, "t")).Get(context.Background(), "app-slug", "")
	require.Error(t, err)
	apiErr, ok := err.(*bitriseapi.APIError)
	require.True(t, ok, "expected *bitriseapi.APIError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestUpdate_ParsesAndSendsYAML(t *testing.T) {
	var gotBody string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/apps/app-slug/bitrise.yml", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{}`))
	})

	err := NewService(bitriseapi.New(srv.URL, "t")).Update(context.Background(), "app-slug", "format_version: \"13\"\n")
	require.NoError(t, err)
	assert.JSONEq(t, `{"app_config_datastore_yaml":{"format_version":"13"}}`, gotBody)
}

func TestUpdate_InvalidYAML(t *testing.T) {
	err := NewService(bitriseapi.New("http://unused.test", "t")).Update(context.Background(), "app-slug", "not: valid: yaml:")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse bitrise.yml")
}

func TestUpdate_AppNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	err := NewService(bitriseapi.New(srv.URL, "t")).Update(context.Background(), "app-slug", "format_version: \"13\"\n")
	require.EqualError(t, err, `app "app-slug" not found`)
}

func TestValidate_Valid(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[],"warnings":[]}`))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Validate(context.Background(), "format_version: \"13\"\n", "")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, []string{}, result.Errors)
	assert.Equal(t, []string{}, result.Warnings)
}

func TestValidate_OmittedSlicesAreNormalized(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Validate(context.Background(), "format_version: \"13\"\n", "")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, []string{}, result.Errors)
	assert.Equal(t, []string{}, result.Warnings)
}

func TestValidate_WithWarnings(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[],"warnings":["deprecated step used"]}`))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Validate(context.Background(), "format_version: \"13\"\n", "app-slug")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, []string{"deprecated step used"}, result.Warnings)
}

func TestValidate_422TreatedAsInvalid(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"could not parse YAML"}`))
	})

	result, err := NewService(bitriseapi.New(srv.URL, "t")).Validate(context.Background(), "not: valid: yaml:", "")
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, []string{"could not parse YAML"}, result.Errors)
	// Both slices are always non-nil, on every return path: ValidateResult's
	// JSON tags have no omitempty, so a nil slice would serialize as null
	// instead of [].
	assert.Equal(t, []string{}, result.Warnings)
}

func TestValidate_PropagatesOtherErrors(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream exploded"}`))
	})

	_, err := NewService(bitriseapi.New(srv.URL, "t")).Validate(context.Background(), "format_version: \"13\"\n", "")
	require.Error(t, err)
	apiErr, ok := err.(*bitriseapi.APIError)
	require.True(t, ok, "expected *bitriseapi.APIError, got %T", err)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
