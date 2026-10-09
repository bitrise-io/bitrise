package build

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestRebuildCmd_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"build-2","build_number":42,"build_url":"https://app.bitrise.io/build/build-2","triggered_workflow":"primary"}`))
	})

	cmd, out := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set(cmdutil.FormatKey, "json"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds/build-1/rebuild", gotPath)
	assert.JSONEq(t, `{}`, string(gotBody))

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "build-2", got["id"])
	assert.Equal(t, float64(42), got["build_number"])
	assert.Equal(t, "primary", got["workflow"])
}

func TestRebuildCmd_RemoteAccess(t *testing.T) {
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"build-2","build_number":42}`))
	})

	cmd, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("remote-access", "true"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.JSONEq(t, `{"is_remote":true}`, string(gotBody))
}

func TestRebuildCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestRebuildCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "--app is required")
}

func TestRebuildCmd_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	err := cmd.RunE(cmd, []string{"missing"})
	require.EqualError(t, err, `build "missing" not found`)
}

func TestRebuildCmd_RejectsWrongArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"a", "b"}} {
		cmd := NewRebuildCommand()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		assert.Error(t, cmd.Execute(), "args=%v should be rejected", args)
	}
}

func newTestRebuildCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewRebuildCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
