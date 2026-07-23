package bitriseapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSteps_PassesAuthHeaderAndQuery(t *testing.T) {
	var gotPath, gotAuth string
	var gotQuery map[string][]string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`[]`))
	})

	_, err := New(srv.URL, "my-token").SearchSteps(context.Background(), StepSearchOptions{
		Query:       "clone",
		Categories:  []string{"utility"},
		Maintainers: []string{"bitrise"},
	})
	require.NoError(t, err)
	assert.Equal(t, "/search-steps", gotPath)
	assert.Equal(t, "token my-token", gotAuth)
	assert.Equal(t, []string{"clone"}, gotQuery["query"])
	assert.Equal(t, []string{"utility"}, gotQuery["categories"])
	assert.Equal(t, []string{"bitrise"}, gotQuery["maintainers"])
}

func TestSearchSteps_ParsesResponse(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","step_ref":"git-clone@8.3.1","title":"Git Clone","maintainer":"bitrise","is_deprecated":false}]`))
	})

	steps, err := New(srv.URL, "t").SearchSteps(context.Background(), StepSearchOptions{Query: "clone"})
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "git-clone@8.3.1", steps[0].StepRef)
	assert.Equal(t, "Git Clone", steps[0].Title)
	assert.Equal(t, "bitrise", steps[0].Maintainer)
}

func TestSearchSteps_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})

	_, err := New(srv.URL, "bad-token").SearchSteps(context.Background(), StepSearchOptions{})
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.Equal(t, "Unauthorized", apiErr.Message)
}

func TestStepInputs_PassesStepRef(t *testing.T) {
	var gotPath, gotStepRef string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotStepRef = r.URL.Query().Get("step_ref")
		_, _ = w.Write([]byte(`[]`))
	})

	_, err := New(srv.URL, "t").StepInputs(context.Background(), "git-clone@8.3.1")
	require.NoError(t, err)
	assert.Equal(t, "/step-inputs", gotPath)
	assert.Equal(t, "git-clone@8.3.1", gotStepRef)
}

func TestStepInputs_ParsesResponse(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"name":"branch","default_value":"main","is_required":true}]`))
	})

	inputs, err := New(srv.URL, "t").StepInputs(context.Background(), "git-clone@8.3.1")
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, "branch", inputs[0].Name)
	assert.Equal(t, "main", inputs[0].DefaultValue)
	assert.True(t, inputs[0].IsRequired)
}

func TestStepInputs_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"step not found"}`))
	})

	_, err := New(srv.URL, "t").StepInputs(context.Background(), "does-not-exist@1.0.0")
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "step not found", apiErr.Message)
}
