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

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestAbortCmd_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	cmd, out := newTestAbortCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("reason", "no longer needed"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds/build-1/abort", gotPath)
	assert.Contains(t, out.String(), "build-1")

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "no longer needed", sent["abort_reason"])
}

func TestAbortCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestAbortCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "--app is required")
}

func TestAbortCmd_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _ := newTestAbortCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	err := cmd.RunE(cmd, []string{"missing"})
	require.EqualError(t, err, `build "missing" not found`)
}

func TestAbortCmd_RejectsWrongArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"a", "b"}} {
		cmd := NewAbortCommand()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		assert.Error(t, cmd.Execute(), "args=%v should be rejected", args)
	}
}

func newTestAbortCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewAbortCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
