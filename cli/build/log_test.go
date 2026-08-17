package build

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestLogCmd_StreamsLogChunks(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"line one\n","position":0},{"chunk":"line two\n","position":1}]}`))
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Equal(t, "/apps/my-app/builds/b-1/log", gotPath)
	assert.Contains(t, out.String(), "line one")
	assert.Contains(t, out.String(), "line two")
}

func TestLogCmd_OrdersOutOfPositionChunks(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"two\n","position":1},{"chunk":"zero\n","position":0},{"chunk":"three\n","position":3},{"chunk":"one\n","position":2}]}`))
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Equal(t, "zero\ntwo\none\nthree\n", out.String())
}

func TestLogCmd_WaitPolls_ThenPrintsLog(t *testing.T) {
	var viewCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1":
			n := int(viewCalls.Add(1))
			if n == 1 {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":1,"triggered_at":"2026-05-06T10:00:00Z"}}`))
			}
		case "/apps/my-app/builds/b-1/log":
			_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"done\n","position":0}]}`))
		}
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Contains(t, out.String(), "Waiting for build")
	assert.Contains(t, out.String(), "done")
}

func TestLogCmd_WaitSkipsPolling_WhenAlreadyFinished(t *testing.T) {
	var viewCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1":
			viewCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":1,"triggered_at":"2026-05-06T10:00:00Z"}}`))
		case "/apps/my-app/builds/b-1/log":
			_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"all good\n","position":0}]}`))
		}
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.NotContains(t, out.String(), "Waiting for build", "should not print wait header for a finished build")
	assert.Equal(t, 1, int(viewCalls.Load()), "expected exactly 1 View call")
	assert.Contains(t, out.String(), "all good")
}

func TestLogCmd_StreamsArchivedLog(t *testing.T) {
	rawSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ARCHIVED LOG CONTENT\n"))
	}))
	t.Cleanup(rawSrv.Close)

	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":true,"expiring_raw_log_url":"` + rawSrv.URL + `"}`))
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Contains(t, out.String(), "ARCHIVED LOG CONTENT")
}

func TestLogCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestLogCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, []string{"b-1"})
	require.EqualError(t, err, "--app is required")
}

func newTestLogCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewLogCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
