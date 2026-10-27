package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
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
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Equal(t, "zero\ntwo\none\nthree\n", out.String())
}

func TestLogCmd_PartialLogNotice(t *testing.T) {
	for _, tt := range []struct {
		name       string
		archived   bool
		flags      map[string]string
		wantNotice bool
	}{
		{"streamed, running", false, nil, true},
		{"tail, running", false, map[string]string{"tail": "1"}, true},
		{"json, running", false, map[string]string{"format": "json"}, false},
		{"streamed, archived", true, nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `{"is_archived":%t,"log_chunks":[{"chunk":"so far\n","position":0}]}`, tt.archived)
			})

			cmd, out := newTestLogCmd(t, srv.URL)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)
			require.NoError(t, cmd.Flags().Set("app", "my-app"))
			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}
			require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

			assert.Contains(t, out.String(), "so far")
			if tt.wantNotice {
				assert.Contains(t, errOut.String(), "The log is not final yet")
			} else {
				assert.Empty(t, errOut.String())
			}
		})
	}
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

func TestLogCmd_Wait_CtrlCDetaches(t *testing.T) {
	// log --wait views the build once before waiting, so Ctrl-C lands on the
	// second View.
	var viewCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apps/my-app/builds/b-1" {
			http.NotFound(w, r)
			return
		}
		if viewCalls.Add(1) == 2 {
			sendInterrupt(t)
		}
		_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`))
	})

	cmd, _ := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1h"))

	err := cmd.RunE(cmd, []string{"b-1"})
	require.EqualError(t, err, "detached: use 'bitrise build log --wait b-1 --interval 1h0m0s --app my-app' to resume")
}

func TestLogCmd_Tail(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":true,"log_chunks":[{"chunk":"one\ntwo\nthree\n","position":0}]}`))
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("tail", "2"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.Equal(t, "two\nthree\n", out.String())
}

func TestLogCmd_FullLogJSON(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"one\ntwo\n","position":0}]}`))
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, []string{"b-1"}))

	assert.JSONEq(t, `{"log":"one\ntwo\n","complete":false}`, out.String())
}

func TestLogCmd_StepLogsByStatus(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step-status", "failed"))
	require.NoError(t, cmd.Flags().Set("tail", "1"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Equal(t, "primary › Tests  failed  1m45s\nTests: line 3\n", out.String())
}

func TestLogCmd_StepLogsByID(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step", "s-1,s-3"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Equal(t, "_setup › Git clone  success  <1s\nGit clone: line 1\nGit clone: line 2\nGit clone: line 3\n\nprimary › save-cache  skipped\n", out.String())
}

func TestLogCmd_StepLogsJSON(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step-status", "failed"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	var got struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
			Log string `json:"log"`
		} `json:"items"`
		Complete bool `json:"complete"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got), out.String())
	assert.True(t, got.Complete)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "s-2", got.Items[0].ID)
	assert.Equal(t, "failed", got.Items[0].Status)
	require.Len(t, got.Items[0].Errors, 1)
	assert.Equal(t, "Tests: line 1\nTests: line 2\nTests: line 3\n", got.Items[0].Log)
}

func TestLogCmd_StepLogsNoMatch(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, out := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step-status", "aborted"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Equal(t, "No steps with status \"aborted\".\n", out.String())
}

func TestLogCmd_StepLogsNoMatchBeforeArchive(t *testing.T) {
	srv := newStepLogsFakeServer(t, `{"is_log_archived": false, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "git-clone", "status": "success"}
]}]}}`)

	cmd, out := newTestLogCmd(t, srv.URL)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step-status", "failed"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.Empty(t, out.String(), "a running build may still fail, so no \"No steps with status\" claim")
	assert.Contains(t, errOut.String(), "The step logs are not available yet")
}

func TestLogCmd_StepLogsUnknownStep(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, _ := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step", "nope"))
	err := cmd.RunE(cmd, []string{"build-1"})
	require.EqualError(t, err, `step "nope" not found in build "build-1"`)
}

func TestLogCmd_StepLogsBeforeArchiveNotice(t *testing.T) {
	srv := newStepLogsFakeServer(t, `{"is_log_archived": false, "execution": {"workflows": [{"workflow_id": "primary", "steps": [
  {"uuid": "s-1", "step_id": "git-clone", "status": "success"},
  {"uuid": "s-2", "step_id": "script", "title": "Tests", "status": "started", "started_at": "2026-10-20T11:56:50Z"}
]}]}}`)

	for format, wantNotice := range map[string]bool{"raw": true, "json": false} {
		cmd, out := newTestLogCmd(t, srv.URL)
		var errOut bytes.Buffer
		cmd.SetErr(&errOut)
		require.NoError(t, cmd.Flags().Set("app", "my-app"))
		require.NoError(t, cmd.Flags().Set("step", "s-1,s-2"))
		require.NoError(t, cmd.Flags().Set("format", format))
		require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

		if wantNotice {
			assert.Equal(t, "primary › git-clone  success\nprimary › Tests  started\n", out.String(), "without logs, no blank line separates the steps")
			assert.Contains(t, errOut.String(), "The step logs are not available yet")
		} else {
			assert.Contains(t, out.String(), `"complete": false`)
			assert.Contains(t, out.String(), `"log": ""`)
			assert.Empty(t, errOut.String())
		}
	}
}

func TestLogCmd_WaitForArchiveBeforeStepLogs(t *testing.T) {
	var summaryCalls atomic.Int32
	steps := newStepLogsFakeServer(t, stepsSummaryJSON)
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/build-1":
			_, _ = w.Write([]byte(`{"data":{"slug":"build-1","build_number":5,"status":2,"triggered_at":"2026-05-06T10:00:00Z"}}`))
		case "/apps/my-app/builds/build-1/log/summary":
			if summaryCalls.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"is_log_archived": false}`))
				return
			}
			steps.Config.Handler.ServeHTTP(w, r)
		default:
			steps.Config.Handler.ServeHTTP(w, r)
		}
	})

	cmd, out := newTestLogCmd(t, srv.URL)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.Flags().Set("step-status", "failed"))
	require.NoError(t, cmd.RunE(cmd, []string{"build-1"}))

	assert.GreaterOrEqual(t, int(summaryCalls.Load()), 3, "polls until archived, then reads the summary for the steps")
	assert.Contains(t, errOut.String(), "Waiting for the build log to be archived")
	assert.NotContains(t, out.String(), "Waiting for the build log", "the notice must not corrupt stdout")
	assert.Contains(t, out.String(), "Tests: line 3")
	assert.NotContains(t, errOut.String(), "not available yet")
}

func TestLogCmd_RejectsUnknownStatus(t *testing.T) {
	srv := newStepLogsFakeServer(t, stepsSummaryJSON)

	cmd, _ := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("step-status", "bogus"))
	require.ErrorContains(t, cmd.RunE(cmd, []string{"build-1"}), `unknown step status "bogus"`)
}

func TestLogCmd_RejectsNegativeTail(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"is_archived":true}`))
	})

	cmd, _ := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("tail", "-1"))
	require.Error(t, cmd.RunE(cmd, []string{"b-1"}))
}

func TestLogCmd_ValidatesFlagsBeforeWaiting(t *testing.T) {
	for name, flags := range map[string]map[string]string{
		"unknown step status": {"step-status": "faild"},
		"negative tail":       {"tail": "-5"},
		"zero tail":           {"tail": "0"},
	} {
		t.Run(name, func(t *testing.T) {
			var calls atomic.Int32
			srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`))
			})

			cmd, _ := newTestLogCmd(t, srv.URL)
			require.NoError(t, cmd.Flags().Set("app", "my-app"))
			require.NoError(t, cmd.Flags().Set("wait", "true"))
			require.NoError(t, cmd.Flags().Set("interval", "1h"))
			for k, v := range flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}
			require.Error(t, cmd.RunE(cmd, []string{"b-1"}))
			assert.Zero(t, calls.Load(), "an invalid flag must fail before the API is called, let alone waited on")
		})
	}
}

func TestLogCmd_Wait_CtrlCResumeHintKeepsFilters(t *testing.T) {
	var viewCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apps/my-app/builds/b-1" {
			http.NotFound(w, r)
			return
		}
		if viewCalls.Add(1) == 2 {
			sendInterrupt(t)
		}
		_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":0,"triggered_at":"2026-05-06T10:00:00Z"}}`))
	})

	cmd, _ := newTestLogCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1h"))
	require.NoError(t, cmd.Flags().Set("step-status", "failed"))
	require.NoError(t, cmd.Flags().Set("step", "s-1,s-2"))
	require.NoError(t, cmd.Flags().Set("tail", "50"))
	require.NoError(t, cmd.Flags().Set("format", "json"))

	err := cmd.RunE(cmd, []string{"b-1"})
	require.EqualError(t, err, "detached: use 'bitrise build log --wait b-1 --step-status failed --step s-1 --step s-2 --tail 50 --interval 1h0m0s --format json --app my-app' to resume")
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

// newStepLogsFakeServer answers build-1's step summary with summary, and the
// log of each step with an archive holding three lines prefixed with the
// step's title, except s-3, which logged nothing.
func newStepLogsFakeServer(t *testing.T, summary string) *httptest.Server {
	t.Helper()
	titles := map[string]string{"s-1": "Git clone", "s-2": "Tests"}
	archive := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		title := titles[strings.TrimPrefix(r.URL.Path, "/")]
		_, _ = fmt.Fprintf(w, `[{"message":"%[1]s: line 1\n"},{"message":"%[1]s: line 2\n"},{"message":"%[1]s: line 3\n"}]`, title)
	})
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if id, ok := strings.CutPrefix(r.URL.Path, "/apps/my-app/builds/build-1/log/steps/"); ok {
			if id == "s-3" {
				_, _ = w.Write([]byte(`{"log_chunks":[]}`))
				return
			}
			_, _ = fmt.Fprintf(w, `{"expiring_raw_log_url":%q}`, archive.URL+"/"+id)
			return
		}
		_, _ = w.Write([]byte(summary))
	})
}
