package session

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestViewCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces/ws-1/sessions/"+uuidSession {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"session":{
			"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
			"templateId":"t-1","sshAddress":"ssh://host:22",
			"templateSnapshot":{"stackId":"osx-xcode-16.0.x-edge","machineType":"g2.mac"}
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"dev", "s-1", "running", "ssh://host:22", "osx-xcode-16.0.x-edge", "g2.mac"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestViewCmd_JSONOutput_MapsStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got["id"] != "s-1" || got["status"] != "running" {
		t.Errorf("unexpected JSON: %v", got)
	}
}

// TestViewCmd_JSONOmitsSSHPassword pins the security contract: the SSH
// password the API returns (consumed internally by `session exec`) must
// never appear in the stable --format json shape. Session.SSHPassword
// carries json:"-".
func TestViewCmd_JSONOmitsSSHPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","sshAddress":"ssh://h:22","sshPassword":"hunter2"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(stdout, "hunter2") || strings.Contains(stdout, "ssh_password") {
		t.Errorf("SSH password leaked into JSON output:\n%s", stdout)
	}
	// ...but the non-secret SSH address is still part of the contract.
	if !strings.Contains(stdout, "ssh_address") {
		t.Errorf("ssh_address missing from JSON:\n%s", stdout)
	}
}

func TestViewCmd_WatchRejectsJSON(t *testing.T) {
	// --watch + --format json must fail fast before any HTTP call (the JSON
	// contract is a single object, not a stream).
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be hit when --watch + json is rejected")
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{"s-1", "--watch"}, Format: output.FormatJSON})
	if err == nil || !strings.Contains(err.Error(), "json") {
		t.Errorf("error = %v, want --watch/json incompatibility", err)
	}
}

// TestViewCmd_WatchRejectsYML is the yml half of TestViewCmd_WatchRejectsJSON:
// the reference implementation only guards against its one non-human format
// (JSON). Here `yml` is a second machine-readable format that would just as
// happily corrupt the redraw loop if it slipped through, so the rejection
// must be format != raw, not format == json.
func TestViewCmd_WatchRejectsYML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be hit when --watch + yml is rejected")
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{"s-1", "--watch"}, Format: output.FormatYML})
	if err == nil || !strings.Contains(err.Error(), "yml") {
		t.Errorf("error = %v, want --watch/yml incompatibility", err)
	}
}

func TestNotificationsCmd_HappyPath_MapsType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces/ws-1/sessions/"+uuidSession+"/notifications" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"notifications":[
			{"id":"n-1","type":"SESSION_NOTIFICATION_TYPE_AGENT_STOPPED","title":"Agent stopped","createdAt":"2026-05-28T10:00:00Z"}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newNotificationsCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"agent_stopped", "Agent stopped", "n-1"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestNotificationsCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"notifications":[
			{"id":"n-1","type":"SESSION_NOTIFICATION_TYPE_AGENT_STOPPED","title":"Agent stopped"}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newNotificationsCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 || got.Items[0].Type != "agent_stopped" {
		t.Errorf("unexpected JSON items: %+v", got.Items)
	}
}

func TestNotificationsCmd_EmptyHuman(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"notifications":[]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newNotificationsCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "No notifications.") {
		t.Errorf("expected empty-state message, got: %q", stdout)
	}
}

func TestNotificationsCmd_RejectsBadOrder(t *testing.T) {
	_, _, err := cmdtest.Run(t, newNotificationsCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"s-1", "--order", "sideways"},
	})
	if err == nil || !strings.Contains(err.Error(), "asc") {
		t.Errorf("error = %v, want --order validation error", err)
	}
}

func TestNotificationsCmd_RejectsBadSince(t *testing.T) {
	_, _, err := cmdtest.Run(t, newNotificationsCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"s-1", "--since", "yesterday"},
	})
	if err == nil || !strings.Contains(err.Error(), "--since") {
		t.Errorf("error = %v, want --since parse error", err)
	}
}
