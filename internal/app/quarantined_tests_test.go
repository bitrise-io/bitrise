package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestListQuarantinedTests_MapsAPIShape(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"quarantined_tests":[{
			"id":"qt-1","test_case_name":"test_login","test_case_identifier":null,"test_suite_name":["LoginTests"],
			"class_name":null,"quarantined_at":"2026-11-04T11:15:00.000+01:00","comment":null,
			"filters":{"branches":null,"targets":["primary"]},
			"quarantined_from":{"number":42,"source":"build","slug":"build-1"},"quarantined_by":"jdoe"
		}]}`))
	})

	res, err := NewService(newAPIClient(t, srv.URL)).ListQuarantinedTests(context.Background(), "my-app")
	require.NoError(t, err)
	assert.Equal(t, QuarantinedTestsResult{Items: []QuarantinedTest{{
		ID:              "qt-1",
		TestCaseName:    "test_login",
		TestSuiteName:   []string{"LoginTests"},
		QuarantinedAt:   time.Date(2026, 11, 4, 10, 15, 0, 0, time.UTC),
		QuarantinedBy:   "jdoe",
		Filters:         QuarantineFilters{Branches: []string{}, Targets: []string{"primary"}},
		QuarantinedFrom: QuarantineSource{Source: "build", ID: "build-1", Number: 42},
	}}}, res)
}

func TestListQuarantinedTests_EmptyIsNotNil(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"quarantined_tests":[]}`))
	})

	res, err := NewService(newAPIClient(t, srv.URL)).ListQuarantinedTests(context.Background(), "my-app")
	require.NoError(t, err)
	b, err := json.Marshal(res)
	require.NoError(t, err)
	assert.JSONEq(t, `{"items":[]}`, string(b))
}

func TestQuarantinedTest_JSONAndYAMLShape(t *testing.T) {
	qt := QuarantinedTest{
		ID:                 "qt-1",
		TestCaseName:       "test_login",
		TestCaseIdentifier: "LoginSpec#test_login",
		TestSuiteName:      []string{"LoginTests"},
		ClassName:          "LoginSpec",
		Comment:            "Flaky",
		QuarantinedAt:      time.Date(2026, 11, 4, 10, 15, 0, 0, time.UTC),
		QuarantinedBy:      "jdoe",
		Filters:            QuarantineFilters{Branches: []string{"main"}, Targets: []string{}},
		QuarantinedFrom:    QuarantineSource{Source: "build", ID: "build-1", Number: 42},
	}

	b, err := json.Marshal(qt)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id":"qt-1","test_case_name":"test_login","test_case_identifier":"LoginSpec#test_login",
		"test_suite_name":["LoginTests"],"class_name":"LoginSpec","comment":"Flaky",
		"quarantined_at":"2026-11-04T10:15:00Z","quarantined_by":"jdoe",
		"filters":{"branches":["main"],"targets":[]},
		"quarantined_from":{"source":"build","id":"build-1","number":42}
	}`, string(b))

	y, err := yaml.Marshal(qt)
	require.NoError(t, err)
	var keys map[string]any
	require.NoError(t, yaml.Unmarshal(y, &keys))
	for _, k := range []string{"id", "test_case_name", "test_case_identifier", "test_suite_name", "class_name", "comment", "quarantined_at", "quarantined_by", "filters", "quarantined_from"} {
		assert.Contains(t, keys, k)
	}
}

func TestAddQuarantinedTest_WithoutBuildSendsFieldsAsGiven(t *testing.T) {
	var gotBody map[string]any
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"qt-1","test_case_name":"test_login","test_suite_name":["LoginTests"],"quarantined_from":{"source":"api"}}`))
	})

	qt, err := NewService(newAPIClient(t, srv.URL)).AddQuarantinedTest(context.Background(), "my-app", AddQuarantinedTestOptions{
		TestCaseName:       "test_login",
		TestSuiteName:      []string{"AppTests", "LoginTests"},
		ClassName:          "LoginSpec",
		TestCaseIdentifier: "LoginSpec#test_login",
	})
	require.NoError(t, err)
	assert.Equal(t, "qt-1", qt.ID)
	assert.Equal(t, map[string]any{
		"test_case_name":       "test_login",
		"test_suite_name":      []any{"AppTests", "LoginTests"},
		"class_name":           "LoginSpec",
		"test_case_identifier": "LoginSpec#test_login",
	}, gotBody, "no filters key when no branch or target is given")
}

func TestAddQuarantinedTest_ConflictNamesExisting(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error_msg":"This test is already quarantined, update its scope instead","quarantined_test":{"id":"qt-existing"}}`))
	})

	_, err := NewService(newAPIClient(t, srv.URL)).AddQuarantinedTest(context.Background(), "my-app", AddQuarantinedTestOptions{TestCaseName: "x", TestSuiteName: []string{"y"}})
	require.Error(t, err)
	assert.Equal(t, "the test is already quarantined as qt-existing; change its scope with 'bitrise app quarantined-tests update qt-existing'", err.Error())
}

// buildResultsJSON nests every test case under a Workflow-named suite that is
// not one of the case's own suites, like real reports do.
const buildResultsJSON = `{"test_suites": [{"name": "primary", "test_cases": [], "test_suites": [
  {"name": "LoginTests", "test_cases": [
    {"name": "testWrongPassword", "class_name": "LoginTests", "status": "failed", "test_suites": ["AppTests", "LoginTests"],
     "properties": [{"key": "test-identifier", "value": "LoginTests/testWrongPassword()"}]},
    {"name": "testWrongPassword", "class_name": "LoginTests", "status": "failed", "test_suites": ["AppTests", "LoginTests"],
     "properties": [{"key": "test-identifier", "value": "LoginTests/testWrongPassword()"}]},
    {"name": "testOK", "class_name": "LoginTests", "status": "passed", "test_suites": ["AppTests", "LoginTests"]},
    {"name": "testOK", "class_name": "SSOTests", "status": "passed", "test_suites": ["AppTests", "SSOTests"]},
    {"name": "testOrphan", "class_name": "LoginTests", "status": "failed"}
  ], "test_suites": []}
]}]}`

func TestAddQuarantinedTest_FromBuild(t *testing.T) {
	tests := []struct {
		name     string
		opts     AddQuarantinedTestOptions
		wantBody map[string]any
	}{
		{
			name: "by name, with the build's values instead of the given ones",
			opts: AddQuarantinedTestOptions{TestCaseName: "testWrongPassword", BuildID: "b-1"},
			wantBody: map[string]any{
				"test_case_name": "testWrongPassword", "class_name": "LoginTests", "test_suite_name": []any{"AppTests", "LoginTests"},
				"test_case_identifier": "LoginTests/testWrongPassword()", "build_slug": "b-1",
			},
		},
		{
			name: "by class name and name as test-results shows them",
			opts: AddQuarantinedTestOptions{TestCaseName: "SSOTests.testOK", BuildID: "b-1"},
			wantBody: map[string]any{
				"test_case_name": "testOK", "class_name": "SSOTests", "test_suite_name": []any{"AppTests", "SSOTests"}, "build_slug": "b-1",
			},
		},
		{
			name: "narrowed by a suite",
			opts: AddQuarantinedTestOptions{TestCaseName: "testOK", TestSuiteName: []string{"LoginTests"}, BuildID: "b-1", Targets: []string{"primary"}},
			wantBody: map[string]any{
				"test_case_name": "testOK", "class_name": "LoginTests", "test_suite_name": []any{"AppTests", "LoginTests"}, "build_slug": "b-1",
				"filters": map[string]any{"branches": nil, "targets": []any{"primary"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					assert.Equal(t, "/apps/my-app/builds/b-1/test-results", r.URL.Path)
					_, _ = w.Write([]byte(buildResultsJSON))
					return
				}
				b, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(b, &gotBody)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"qt-1"}`))
			})

			_, err := NewService(newAPIClient(t, srv.URL)).AddQuarantinedTest(context.Background(), "my-app", tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestAddQuarantinedTest_FromBuildRefusesWithoutOneMatch(t *testing.T) {
	tests := []struct {
		name    string
		opts    AddQuarantinedTestOptions
		wantErr []string
	}{
		{
			name:    "no match",
			opts:    AddQuarantinedTestOptions{TestCaseName: "testMissing", BuildID: "b-1"},
			wantErr: []string{`build b-1 has no test case "testMissing"`, "'bitrise build test-results b-1'"},
		},
		{
			name:    "a flag rules out the only case with the name",
			opts:    AddQuarantinedTestOptions{TestCaseName: "testWrongPassword", TestCaseIdentifier: "other", BuildID: "b-1"},
			wantErr: []string{`no test case "testWrongPassword"`},
		},
		{
			name: "ambiguous",
			opts: AddQuarantinedTestOptions{TestCaseName: "testOK", BuildID: "b-1"},
			wantErr: []string{
				"2 test cases of build b-1 match; narrow it down with --class-name, --suite or --identifier:",
				"suites: AppTests › LoginTests  class name: LoginTests  name: testOK",
				"suites: AppTests › SSOTests  class name: SSOTests  name: testOK",
			},
		},
		{
			name:    "no suites",
			opts:    AddQuarantinedTestOptions{TestCaseName: "testOrphan", BuildID: "b-1"},
			wantErr: []string{`lists no suites for test case "testOrphan", so it cannot be quarantined`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Error("no quarantine expected")
				}
				_, _ = w.Write([]byte(buildResultsJSON))
			})

			_, err := NewService(newAPIClient(t, srv.URL)).AddQuarantinedTest(context.Background(), "my-app", tt.opts)
			require.Error(t, err)
			for _, want := range tt.wantErr {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

func TestAddQuarantinedTest_FromBuildPassesTestResultsErrors(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_msg":"Test reports are not available on the plan of this workspace"}`))
	})

	_, err := NewService(newAPIClient(t, srv.URL)).AddQuarantinedTest(context.Background(), "my-app", AddQuarantinedTestOptions{TestCaseName: "x", BuildID: "b-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting the test results of build b-1")
	assert.Contains(t, err.Error(), "does not include test reports")
}
