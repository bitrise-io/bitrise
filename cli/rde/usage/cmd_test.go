package usage

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/output"
)

// sparseReport mimics the backend's proto3 JSON: zero-valued fields and empty
// bucket objects omitted. Alice runs 1 linux (2 vCPU / 8 GB) + 1 macos
// (4 / 6); the workspace bucket runs 1 session of unknown machine type.
const sparseReport = `{
	"totals": {
		"linux": {"sessionCount": 1, "vcpu": 2, "memoryGb": 8},
		"macos": {"sessionCount": 1, "vcpu": 4, "memoryGb": 6},
		"unknown": {"sessionCount": 1}
	},
	"users": [
		{
			"userId": "u-1", "userSlug": "alice-slug", "email": "alice@example.com", "username": "alice",
			"totals": {
				"linux": {"sessionCount": 1, "vcpu": 2, "memoryGb": 8},
				"macos": {"sessionCount": 1, "vcpu": 4, "memoryGb": 6}
			}
		},
		{
			"isWorkspace": true,
			"totals": {"unknown": {"sessionCount": 1}}
		}
	],
	"unknownMachineTypeCount": 1
}`

func TestUsageCmd_Human(t *testing.T) {
	srv := usageServer(t, http.StatusOK, sparseReport)
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"Sessions:", "3", "(Linux 1, macOS 1, unknown 1)",
		"vCPU:", "6", "(Linux 2, macOS 4)",
		"Memory GB:", "14", "(Linux 8, macOS 6)",
		"alice@example.com", "(workspace)",
		"2 / 8", "4 / 6",
		"UNKNOWN VCPU/GB",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "Warning") {
		t.Errorf("diagnostics leaked into stdout:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Warning:") || !strings.Contains(stderr, "(1 affected)") {
		t.Errorf("stderr missing the undercount warning:\n%s", stderr)
	}
}

func TestUsageCmd_JSON_Densified(t *testing.T) {
	srv := usageServer(t, http.StatusOK, sparseReport)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}

	totals, ok := got["totals"].(map[string]any)
	if !ok {
		t.Fatalf("totals missing: %s", stdout)
	}
	// The wire omitted vcpu/memoryGb on the unknown bucket; the stable shape
	// promises them as explicit zeros.
	unknown, ok := totals["unknown"].(map[string]any)
	if !ok {
		t.Fatalf("totals.unknown missing: %s", stdout)
	}
	for key, want := range map[string]float64{"session_count": 1, "vcpu": 0, "memory_gb": 0} {
		if unknown[key] != want {
			t.Errorf("totals.unknown[%q] = %v, want %v", key, unknown[key], want)
		}
	}

	users, ok := got["users"].([]any)
	if !ok || len(users) != 2 {
		t.Fatalf("users = %v, want 2 rows", got["users"])
	}
	// The stable shape's user_id is the Bitrise user ID (the wire's slug);
	// the backend's internal UUID (wire userId) is not exposed.
	alice := users[0].(map[string]any)
	if alice["user_id"] != "alice-slug" {
		t.Errorf("users[0].user_id = %v, want alice-slug", alice["user_id"])
	}
	if _, leaked := alice["user_slug"]; leaked {
		t.Errorf("users[0] leaks user_slug: %v", alice)
	}
	if got["unknown_machine_type_count"] != float64(1) {
		t.Errorf("unknown_machine_type_count = %v, want 1", got["unknown_machine_type_count"])
	}
}

func TestUsageCmd_Human_NoUnknownColumnWhenAllKnown(t *testing.T) {
	srv := usageServer(t, http.StatusOK, `{
		"totals": {"linux": {"sessionCount": 1, "vcpu": 2, "memoryGb": 8}},
		"users": [{
			"userSlug": "alice-slug", "email": "alice@example.com",
			"totals": {"linux": {"sessionCount": 1, "vcpu": 2, "memoryGb": 8}}
		}]
	}`)
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdout, "UNKNOWN") || strings.Contains(stdout, "unknown") {
		t.Errorf("unknown column/split rendered with nothing unknown:\n%s", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want no diagnostics", stderr)
	}
}

func TestUsageCmd_Empty(t *testing.T) {
	srv := usageServer(t, http.StatusOK, `{}`)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No active sessions.") {
		t.Errorf("stdout = %q, want the empty-report sentence", stdout)
	}
}

func TestUsageCmd_PermissionDenied(t *testing.T) {
	srv := usageServer(t, http.StatusForbidden, `{"code":7,"message":"you are not allowed to view usage of this workspace"}`)
	defer srv.Close()

	_, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err == nil || !strings.Contains(err.Error(), "not allowed to view usage") {
		t.Fatalf("error = %v, want the backend's permission message", err)
	}
}

// TestUsageCmd_QuietKeepsWarning attaches the persistent --quiet flag to the
// command under test (cmdutil.IsQuiet reads it off cmd.Root(), which for a
// standalone command is the command itself): the undercount warning survives
// -q, per the output scheme's rule that warnings ignore --quiet.
func TestUsageCmd_QuietKeepsWarning(t *testing.T) {
	srv := usageServer(t, http.StatusOK, sparseReport)
	defer srv.Close()

	cmd := NewCmd()
	cmd.PersistentFlags().BoolP(cmdutil.FlagQuiet, "q", false, "quiet")

	stdout, stderr, err := cmdtest.Run(t, cmd, cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"-q"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Warning:") || !strings.Contains(stderr, "(1 affected)") {
		t.Errorf("-q must not suppress the undercount warning:\n%s", stderr)
	}
	if !strings.Contains(stdout, "Sessions:") {
		t.Errorf("-q must not affect the report itself:\n%s", stdout)
	}
}

// usageServer serves GET /v1/workspaces/ws-1/usage with a fixed body/status.
func usageServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces/ws-1/usage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
}
