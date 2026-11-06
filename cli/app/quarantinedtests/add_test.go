package quarantinedtests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_WithoutBuildSendsFlags(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(quarantinedTestJSON))
	})

	out, _, err := runCmd(t, NewAddCommand(), srv.URL,
		"--name", "test_login",
		"--suite", "AppTests", "--suite", "Login, parameterized",
		"--class-name", "LoginSpec",
		"--identifier", "LoginSpec#test_login",
		"--comment", "Flaky on CI",
		"--branch", "main,release/*",
		"--target", "primary",
	)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/quarantined-tests", gotPath)
	assert.Equal(t, map[string]any{
		"test_case_name":       "test_login",
		"test_suite_name":      []any{"AppTests", "Login, parameterized"},
		"class_name":           "LoginSpec",
		"test_case_identifier": "LoginSpec#test_login",
		"comment":              "Flaky on CI",
		"filters":              map[string]any{"branches": []any{"main", "release/*"}, "targets": []any{"primary"}},
	}, gotBody)
	assert.Contains(t, out, "ID:           qt-1\n")
}

func TestAdd_FromBuildSendsTheBuildsValues(t *testing.T) {
	var requests []string
	var gotBody map[string]any
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"test_suites": [{"name": "primary", "test_cases": [], "test_suites": [{"name": "LoginTests", "test_suites": [], "test_cases": [
				{"name": "test_login", "class_name": "LoginSpec", "status": "failed", "test_suites": ["AppTests", "LoginTests"],
				 "properties": [{"key": "test-identifier", "value": "LoginSpec/test_login()"}]}]}]}]}`))
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(quarantinedTestJSON))
	})

	_, _, err := runCmd(t, NewAddCommand(), srv.URL, "--build", "build-1", "--name", "LoginSpec.test_login", "--comment", "Flaky on CI")
	require.NoError(t, err)
	assert.Equal(t, []string{"GET /apps/my-app/builds/build-1/test-results", "POST /apps/my-app/quarantined-tests"}, requests)
	assert.Equal(t, map[string]any{
		"test_case_name":       "test_login",
		"test_suite_name":      []any{"AppTests", "LoginTests"},
		"class_name":           "LoginSpec",
		"test_case_identifier": "LoginSpec/test_login()",
		"build_slug":           "build-1",
		"comment":              "Flaky on CI",
	}, gotBody, "no targets unless --target is passed")
}

func TestAdd_RequiresNameAndBuildOrSuite(t *testing.T) {
	srv := newFakeServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request expected")
	})

	_, _, err := runCmd(t, NewAddCommand(), srv.URL, "--name", "test_login")
	assert.ErrorContains(t, err, "pass --build")

	_, _, err = runCmd(t, NewAddCommand(), srv.URL, "--suite", "LoginTests")
	assert.ErrorContains(t, err, `"name"`)
}

func TestAdd_AlreadyQuarantined(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error_msg":"This test is already quarantined, update its scope instead","quarantined_test":` + quarantinedTestJSON + `}`))
	})

	_, _, err := runCmd(t, NewAddCommand(), srv.URL, "--name", "test_login", "--suite", "LoginTests")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already quarantined as qt-1")
	assert.Contains(t, err.Error(), "bitrise app quarantined-tests update qt-1")
}

func TestAdd_Forbidden(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_msg":"Test quarantine is not available on the plan of this workspace"}`))
	})

	_, _, err := runCmd(t, NewAddCommand(), srv.URL, "--name", "test_login", "--suite", "LoginTests")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quarantining test failed")
	assert.Contains(t, err.Error(), "Test quarantine is not available on the plan of this workspace")
}
