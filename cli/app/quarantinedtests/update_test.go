package quarantinedtests

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate_SendsOnlyChangedFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "comment", args: []string{"--comment", "Tracked"}, want: `{"comment":"Tracked"}`},
		{name: "empty comment", args: []string{"--comment", ""}, want: `{"comment":""}`},
		{name: "branches", args: []string{"--branch", "main", "--branch", "release/*"}, want: `{"filters":{"branches":["main","release/*"]}}`},
		{name: "all targets", args: []string{"--all-targets"}, want: `{"filters":{"targets":null}}`},
		{name: "all branches and targets", args: []string{"--all-branches", "--all-targets"}, want: `{"filters":{"branches":null,"targets":null}}`},
		{name: "targets and comment", args: []string{"--target", "primary", "--comment", "x"}, want: `{"comment":"x","filters":{"targets":["primary"]}}`},
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

			out, _, err := runCmd(t, NewUpdateCommand(), srv.URL, append([]string{"qt-1"}, tt.args...)...)
			require.NoError(t, err)
			assert.Equal(t, http.MethodPatch, gotMethod)
			assert.Equal(t, "/apps/my-app/quarantined-tests/qt-1", gotPath)
			assert.JSONEq(t, tt.want, gotBody)
			assert.Contains(t, out, "ID:           qt-1\n")
		})
	}
}

func TestUpdate_RejectsInvalidFlags(t *testing.T) {
	tests := map[string][]string{
		"nothing to update":       {"qt-1"},
		"all-branches=false only": {"qt-1", "--all-branches=false"},
		"branch and all-branches": {"qt-1", "--branch", "main", "--all-branches"},
		"target and all-targets":  {"qt-1", "--target", "primary", "--all-targets"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			srv := newFakeServer(t, func(_ http.ResponseWriter, _ *http.Request) {
				t.Error("no request expected")
			})

			_, _, err := runCmd(t, NewUpdateCommand(), srv.URL, args...)
			assert.Error(t, err)
		})
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_msg":"Running builds permission is required to change quarantined tests"}`))
	})

	_, _, err := runCmd(t, NewUpdateCommand(), srv.URL, "qt-1", "--comment", "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating quarantined test failed")
	assert.Contains(t, err.Error(), "Running builds permission is required")
}
