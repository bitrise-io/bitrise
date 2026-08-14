package build

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/log"
)

func TestTriggerCmd_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"build_slug":"new-1","build_number":100,"build_url":"https://app.bitrise.io/build/new-1","triggered_workflow":"primary"}`)
	})

	cmd, out := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/apps/my-app/builds", gotPath)
	assert.Contains(t, out.String(), "Build triggered")
}

func TestTriggerCmd_JSONOutput(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"build_slug":"new-1","build_number":100,"build_url":"https://app.bitrise.io/build/new-1","triggered_workflow":"primary"}`)
	})

	var logBuf strings.Builder
	log.InitGlobalLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Producer: log.BitriseCLI, Writer: &logBuf})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(logBuf.String()), &got))
	assert.Equal(t, "new-1", got["id"])
	assert.Equal(t, float64(100), got["build_number"])
}

func TestTriggerCmd_InvalidEnvJSON(t *testing.T) {
	srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("env", "not-json"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env")
}

func TestTriggerCmd_WorkflowAndPipelineMutuallyExclusive(t *testing.T) {
	cmd, _ := newTestTriggerCmd(t, "https://unused.test")
	cmd.SetArgs([]string{"--app", "my-app", "--workflow", "primary", "--pipeline", "my-pipeline"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestTriggerCmd_WaitAndWatchMutuallyExclusive(t *testing.T) {
	cmd, _ := newTestTriggerCmd(t, "https://unused.test")
	cmd.SetArgs([]string{"--app", "my-app", "--workflow", "primary", "--wait", "--watch"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestTriggerCmd_Wait_BlocksAndExits(t *testing.T) {
	// Two View calls after trigger: first in-progress, then success.
	var viewCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apps/my-app/builds" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"build_slug":"b-1","build_number":5,"build_url":"https://app.bitrise.io/build/b-1","triggered_workflow":"primary"}`)
		case r.URL.Path == "/apps/my-app/builds/b-1":
			n := int(viewCalls.Add(1))
			if n == 1 {
				_, _ = io.WriteString(w, `{"data":{"slug":"b-1","build_number":5,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`)
			} else {
				_, _ = io.WriteString(w, `{"data":{"slug":"b-1","build_number":5,"status":1,"triggered_at":"2026-05-06T10:00:00Z"}}`)
			}
		}
	})

	cmd, out := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, out.String(), "Waiting for build")
	assert.Contains(t, out.String(), "finished")
}

func TestTriggerCmd_Wait_FailedBuildReturnsError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apps/my-app/builds" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"build_slug":"b-1","build_number":5}`)
		case r.URL.Path == "/apps/my-app/builds/b-1":
			_, _ = io.WriteString(w, `{"data":{"slug":"b-1","build_number":5,"status":2,"triggered_at":"2026-05-06T10:00:00Z"}}`)
		}
	})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestTriggerCmd_Wait_FailedBuildJSONWritesRecordAndErrors(t *testing.T) {
	// Regression: with --format json a failed build must still write the
	// build record to stdout AND return a non-zero error so CI scripts can
	// gate on it.
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apps/my-app/builds" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"build_slug":"b-1","build_number":5}`)
		case r.URL.Path == "/apps/my-app/builds/b-1":
			_, _ = io.WriteString(w, `{"data":{"slug":"b-1","build_number":5,"status":2,"triggered_at":"2026-05-06T10:00:00Z"}}`)
		}
	})

	var logBuf strings.Builder
	log.InitGlobalLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Producer: log.BitriseCLI, Writer: &logBuf})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.Flags().Set("format", "json"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")

	var rec map[string]any
	require.NoError(t, json.Unmarshal([]byte(logBuf.String()), &rec))
	assert.Equal(t, "b-1", rec["id"])
}

func TestTriggerCmd_DefaultsBranchToMain(t *testing.T) {
	var gotBranch string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if bp, ok := body["build_params"].(map[string]any); ok {
			gotBranch, _ = bp["branch"].(string)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"build_slug":"x","build_number":1}`)
	})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "main", gotBranch)
}

func TestTriggerCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestTriggerCmd(t, "https://unused.test")
	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, "--app is required")
}

func TestTriggerCmd_AppNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "missing-app"))
	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, `app "missing-app" not found`)
}

func newTestTriggerCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewTriggerCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
