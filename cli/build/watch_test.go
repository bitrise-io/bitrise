package build

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// watchStubServer serves a finished+archived build so Service.Watch takes the
// archived branch (stream log, then return the final View) without polling.
func watchStubServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ARCHIVED LOG LINE\n"))
	}))
	t.Cleanup(raw.Close)

	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			_, _ = w.Write([]byte(`{"is_archived":true,"expiring_raw_log_url":"` + raw.URL + `"}`))
		case "/apps/my-app/builds/b-1":
			_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":7,"status":` + strconv.Itoa(status) + `,"triggered_at":"2026-05-06T10:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	})
	return srv
}

func TestWatchCmd_JSONWritesRecordToStdoutLogsToStderr(t *testing.T) {
	srv := watchStubServer(t, 1) // success

	cmd, stdout, stderr := newTestWatchCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	var rec map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rec))
	assert.Equal(t, "success", rec["status"])
	assert.Contains(t, stderr.String(), "ARCHIVED LOG LINE")
	assert.NotContains(t, stdout.String(), "ARCHIVED LOG LINE", "log lines must not leak into the JSON record")
}

func TestWatchCmd_JSONFailedBuildExitsNonZero(t *testing.T) {
	srv := watchStubServer(t, 2) // failed

	cmd, stdout, _ := newTestWatchCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("format", "json"))

	err := cmd.RunE(cmd, []string{"b-1"})
	require.Error(t, err, "expected non-zero exit (error) for a failed build")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rec))
	assert.Equal(t, "failed", rec["status"])
}

func TestWatchCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _, _ := newTestWatchCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, []string{"b-1"})
	require.EqualError(t, err, "--app is required")
}

func newTestWatchCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewWatchCommand()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &stdout, &stderr
}
