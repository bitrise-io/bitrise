package bitriseapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const quarantinedTestJSON = `{
	"id":"qt-1","test_case_name":"test_login","test_case_identifier":null,"test_suite_name":["AppTests","LoginTests"],
	"class_name":"LoginSpec","quarantined_at":"2026-11-04T10:15:00.000Z","comment":"Flaky",
	"filters":{"branches":["main"],"targets":null},
	"quarantined_from":{"number":42,"source":"build","slug":"build-1"},"quarantined_by":"jdoe"
}`

func TestQuarantinedTests_DecodesList(t *testing.T) {
	var gotMethod, gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"quarantined_tests":[` + quarantinedTestJSON + `]}`))
	})

	tests, err := newAPIClient(t, srv.URL, "t").QuarantinedTests(context.Background(), "my-app")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/apps/my-app/quarantined-tests", gotPath)
	require.Len(t, tests, 1)
	assert.Equal(t, QuarantinedTest{
		ID:              "qt-1",
		TestCaseName:    "test_login",
		TestSuiteName:   []string{"AppTests", "LoginTests"},
		ClassName:       "LoginSpec",
		Comment:         "Flaky",
		QuarantinedAt:   time.Date(2026, 11, 4, 10, 15, 0, 0, time.UTC),
		QuarantinedBy:   "jdoe",
		Filters:         QuarantinedTestFilters{Branches: []string{"main"}},
		QuarantinedFrom: QuarantinedTestSource{Source: "build", Slug: "build-1", Number: 42},
	}, tests[0])
}

func TestQuarantinedTest_GetsByID(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(quarantinedTestJSON))
	})

	qt, err := newAPIClient(t, srv.URL, "t").QuarantinedTest(context.Background(), "my-app", "qt-1")
	require.NoError(t, err)
	assert.Equal(t, "/apps/my-app/quarantined-tests/qt-1", gotPath)
	assert.Equal(t, "qt-1", qt.ID)
}

func TestCreateQuarantinedTest_SendsBody(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody map[string]any
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotContentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(quarantinedTestJSON))
	})

	_, err := newAPIClient(t, srv.URL, "t").CreateQuarantinedTest(context.Background(), "my-app", CreateQuarantinedTestRequest{
		TestCaseName:  "test_login",
		TestSuiteName: []string{"AppTests", "LoginTests"},
		BuildSlug:     "build-1",
		Filters:       &QuarantinedTestFilters{Branches: []string{"main"}},
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/quarantined-tests", gotPath)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, map[string]any{
		"test_case_name":  "test_login",
		"test_suite_name": []any{"AppTests", "LoginTests"},
		"build_slug":      "build-1",
		"filters":         map[string]any{"branches": []any{"main"}, "targets": nil},
	}, gotBody)
}

func TestCreateQuarantinedTest_ConflictCarriesExisting(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error_msg":"This test is already quarantined, update its scope instead","quarantined_test":` + quarantinedTestJSON + `}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").CreateQuarantinedTest(context.Background(), "my-app", CreateQuarantinedTestRequest{TestCaseName: "x", TestSuiteName: []string{"y"}})

	var conflict *AlreadyQuarantinedError
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, "qt-1", conflict.Existing.ID)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
}

func TestCreateQuarantinedTest_OtherErrorsStayAPIErrors(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_msg":"Running builds permission is required to change quarantined tests"}`))
	})

	_, err := newAPIClient(t, srv.URL, "t").CreateQuarantinedTest(context.Background(), "my-app", CreateQuarantinedTestRequest{TestCaseName: "x", TestSuiteName: []string{"y"}})

	var conflict *AlreadyQuarantinedError
	assert.NotErrorAs(t, err, &conflict)
	assert.ErrorContains(t, err, "Running builds permission is required")
}

func TestUpdateQuarantinedTest_SendsOnlySetKeys(t *testing.T) {
	comment := "Tracked"
	branches := []string{"main"}
	var cleared []string

	tests := []struct {
		name string
		req  UpdateQuarantinedTestRequest
		want string
	}{
		{name: "comment only", req: UpdateQuarantinedTestRequest{Comment: &comment}, want: `{"comment":"Tracked"}`},
		{name: "branches only", req: UpdateQuarantinedTestRequest{Branches: &branches}, want: `{"filters":{"branches":["main"]}}`},
		{name: "cleared targets", req: UpdateQuarantinedTestRequest{Targets: &cleared}, want: `{"filters":{"targets":null}}`},
		{name: "nothing", req: UpdateQuarantinedTestRequest{}, want: `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath, gotBody string
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath = r.Method, r.URL.Path
				b, _ := io.ReadAll(r.Body)
				gotBody = string(b)
				_, _ = w.Write([]byte(quarantinedTestJSON))
			})

			_, err := newAPIClient(t, srv.URL, "t").UpdateQuarantinedTest(context.Background(), "my-app", "qt-1", tt.req)
			require.NoError(t, err)
			assert.Equal(t, http.MethodPatch, gotMethod)
			assert.Equal(t, "/apps/my-app/quarantined-tests/qt-1", gotPath)
			assert.JSONEq(t, tt.want, gotBody)
		})
	}
}

func TestDeleteQuarantinedTest_SendsNoBody(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotContentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	require.NoError(t, newAPIClient(t, srv.URL, "t").DeleteQuarantinedTest(context.Background(), "my-app", "qt-1"))
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/apps/my-app/quarantined-tests/qt-1", gotPath)
	assert.Empty(t, gotContentType)
	assert.Empty(t, gotBody)
}
