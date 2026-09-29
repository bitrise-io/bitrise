package build

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
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

	cmd, out := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
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

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("workflow", "primary"))
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("interval", "1ms"))
	require.NoError(t, cmd.Flags().Set("format", "json"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rec))
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

func TestTriggerCmd_CommitHashDoesNotDefaultBranch(t *testing.T) {
	var bp map[string]any
	srv := newBodyCapturingServer(t, &bp)

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, cmd.Flags().Set("commit-hash", "abc123"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "abc123", bp["commit_hash"])
	assert.NotContains(t, bp, "branch")
}

func TestTriggerCmd_Priority(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantSent  bool
		wantValue float64
	}{
		{name: "omitted", args: nil, wantSent: false},
		{name: "explicit zero", args: []string{"--priority", "0"}, wantSent: true, wantValue: 0},
		{name: "high", args: []string{"--priority", "50"}, wantSent: true, wantValue: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bp map[string]any
			srv := newBodyCapturingServer(t, &bp)

			cmd, _ := newTestTriggerCmd(t, srv.URL)
			cmd.SetArgs(append([]string{"--app", "my-app", "--workflow", "primary"}, tt.args...))
			require.NoError(t, cmd.Execute())

			got, ok := bp["priority"]
			require.Equal(t, tt.wantSent, ok)
			if tt.wantSent {
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestTriggerCmd_EnvIsExpand(t *testing.T) {
	var bp map[string]any
	srv := newBodyCapturingServer(t, &bp)

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	cmd.SetArgs([]string{"--app", "my-app", "--workflow", "primary",
		"--env", `{"API_URL":"https://example.com","PRICE":{"value":"$5","is_expand":false},"HOME_DIR":{"value":"$HOME"}}`})
	require.NoError(t, cmd.Execute())

	envs, _ := bp["environments"].([]any)
	require.Len(t, envs, 3)
	got := map[string][2]any{}
	for _, e := range envs {
		env, _ := e.(map[string]any)
		key, _ := env["mapped_to"].(string)
		got[key] = [2]any{env["value"], env["is_expand"]}
	}
	assert.Equal(t, map[string][2]any{
		"API_URL":  {"https://example.com", true},
		"PRICE":    {"$5", false},
		"HOME_DIR": {"$HOME", true},
	}, got)
}

func TestTriggerCmd_EnvInvalidValue(t *testing.T) {
	tests := map[string]string{
		"number":        `{"A":1}`,
		"missing value": `{"A":{"is_expand":false}}`,
		"unknown field": `{"A":{"value":"x","is_sensitive":true}}`,
	}
	for name, envJSON := range tests {
		t.Run(name, func(t *testing.T) {
			srv := newFakeServer(t, func(http.ResponseWriter, *http.Request) {
				t.Error("no request expected")
			})

			cmd, _ := newTestTriggerCmd(t, srv.URL)
			cmd.SetArgs([]string{"--app", "my-app", "--env", envJSON})

			err := cmd.Execute()
			require.EqualError(t, err, `--env: "A" must be a string or {"value":"...","is_expand":false}`)
		})
	}
}

func TestTriggerCmd_MachineFlags(t *testing.T) {
	var bp map[string]any
	srv := newBodyCapturingServer(t, &bp)

	cmd, _ := newTestTriggerCmd(t, srv.URL)
	cmd.SetArgs([]string{"--app", "my-app", "--workflow", "primary",
		"--stack", "osx-xcode-16.0.x", "--machine-type", "g2-m1.4core", "--license-pool", "pool-1"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "osx-xcode-16.0.x", bp["stack"])
	assert.Equal(t, "g2-m1.4core", bp["machine_type_id"])
	assert.Equal(t, "pool-1", bp["license_pool_id"])
}

func TestTriggerCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
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

func newBodyCapturingServer(t *testing.T, bp *map[string]any) *httptest.Server {
	t.Helper()
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			return
		}
		var body struct {
			BuildParams map[string]any `json:"build_params"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		*bp = body.BuildParams
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"build_slug":"x","build_number":1}`)
	})
}
