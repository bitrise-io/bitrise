package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

	cmd, out, _ := newTestRebuildCmd(t, srv.URL)
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

	cmd, _, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("remote-access", "true"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.JSONEq(t, `{"is_remote":true}`, string(gotBody))
}

func TestRebuildCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _, _ := newTestRebuildCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "--app is required")
}

func TestRebuildCmd_BuildNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	err := cmd.RunE(cmd, []string{"missing"})
	require.EqualError(t, err, `build "missing" not found`)
}

func TestRebuildCmd_TextOutput(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"build_slug":"build-2","build_number":42,"build_url":"https://app.bitrise.io/build/build-2","triggered_workflow":"primary"}`))
	})

	cmd, out, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Contains(t, out.String(), "Rebuild started")
	assert.Contains(t, out.String(), "#42")
	assert.Contains(t, out.String(), "https://app.bitrise.io/build/build-2")
}

func TestRebuildCmd_WaitAndWatchMutuallyExclusive(t *testing.T) {
	cmd, _, _ := newTestRebuildCmd(t, "https://unused.test")
	cmd.SetArgs([]string{"build-1", "--app", "my-app", "--wait", "--watch"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestRebuildCmd_Wait_FailedBuildJSONWritesRecordAndErrors(t *testing.T) {
	srv := rebuildStubServer(t, 2) // failed

	cmd, stdout, stderr := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.Flags().Set(cmdutil.FormatKey, "json"))

	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "build failed")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rec))
	assert.Equal(t, "b-2", rec["id"])
	assert.Equal(t, "failed", rec["status"])
	assert.NotContains(t, stderr.String(), "Waiting for build", "the human-only header is raw-format only")
	assert.NotContains(t, stderr.String(), "ARCHIVED LOG LINE", "--wait must not stream the log")
}

func TestRebuildCmd_Watch_StreamsLogAndExits(t *testing.T) {
	srv := rebuildStubServer(t, 1) // success

	cmd, stdout, stderr := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("watch", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Contains(t, stdout.String(), "ARCHIVED LOG LINE")
	assert.Contains(t, stderr.String(), "Watching build #8")
	assert.Contains(t, stderr.String(), "Build #8 finished: success")
}

func TestRebuildCmd_Wait_CtrlCDetaches(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/build-1/rebuild":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"build_slug":"b-2","build_number":8}`))
		case "/apps/my-app/builds/b-2":
			sendInterrupt(t)
			_, _ = w.Write([]byte(`{"data":{"slug":"b-2","build_number":8,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	})

	cmd, _, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1h"))

	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "detached: use 'bitrise build watch b-2 --app my-app' to resume")
}

func TestRebuildCmd_Watch_CtrlCDetaches(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/build-1/rebuild":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"build_slug":"b-2","build_number":8}`))
		case "/apps/my-app/builds/b-2/log":
			sendInterrupt(t)
			_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[]}`))
		default:
			http.NotFound(w, r)
		}
	})

	cmd, _, _ := newTestRebuildCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("watch", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1h"))

	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, "detached: use 'bitrise build watch b-2 --app my-app' to resume")
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

func newTestRebuildCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewRebuildCommand()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &stdout, &stderr
}

// sendInterrupt delivers a real SIGINT to the test process, as Ctrl-C would.
// Callers send it while the command is waiting, so its signal handler is
// installed; otherwise the default handling would kill the test binary. It
// runs on server goroutines, so it reports with Errorf rather than FailNow.
func sendInterrupt(t *testing.T) {
	t.Helper()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Errorf("find own process: %v", err)
		return
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Errorf("send SIGINT: %v", err)
	}
}

// rebuildStubServer rebuilds build-1 into b-2 and serves b-2 as finished and
// archived, so both --wait and --watch complete without polling.
func rebuildStubServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ARCHIVED LOG LINE\n"))
	}))
	t.Cleanup(raw.Close)

	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/build-1/rebuild":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"build_slug":"b-2","build_number":8,"triggered_workflow":"primary"}`))
		case "/apps/my-app/builds/b-2/log":
			_, _ = w.Write([]byte(`{"is_archived":true,"expiring_raw_log_url":"` + raw.URL + `"}`))
		case "/apps/my-app/builds/b-2":
			_, _ = fmt.Fprintf(w, `{"data":{"slug":"b-2","build_number":8,"status":%d,"triggered_at":"2026-05-06T10:00:00Z"}}`, status)
		default:
			http.NotFound(w, r)
		}
	})
}
