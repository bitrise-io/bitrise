package bitriseapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestNew_DefaultHTTPClientTimeout(t *testing.T) {
	c := New("http://example.invalid", "t")
	assert.Equal(t, defaultTimeout, c.httpClient.Timeout)
}

func TestWithHTTPClient_Overrides(t *testing.T) {
	custom := &http.Client{}
	c := New("http://example.invalid", "t", WithHTTPClient(custom))
	assert.Same(t, custom, c.httpClient)
}

func TestAPIError_NonJSONBody(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream exploded"))
	})

	_, err := New(srv.URL, "t").SearchSteps(context.Background(), StepSearchOptions{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, "upstream exploded", apiErr.Body)
	assert.Contains(t, err.Error(), "upstream exploded")
}

func TestAPIError_AlternativeJSONFields(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"error_msg": {`{"error_msg":"bad request"}`, "bad request"},
		"error":     {`{"error":"forbidden"}`, "forbidden"},
		"errors":    {`{"errors":["a is invalid","b is required"]}`, "a is invalid; b is required"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := New(srv.URL, "t").SearchSteps(context.Background(), StepSearchOptions{})
			apiErr, ok := err.(*APIError)
			require.True(t, ok)
			assert.Equal(t, tc.want, apiErr.Message)
		})
	}
}

func TestAPIError_RequestInfo(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})

	_, err := New(srv.URL, "bad-token").SearchSteps(context.Background(), StepSearchOptions{Query: "clone"})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.Contains(t, apiErr.RequestInfo, "GET /search-steps")
	assert.Contains(t, err.Error(), "Unauthorized")
}
