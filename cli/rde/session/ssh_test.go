package session

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

const sshSessionJSON = `{"session":{
	"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING",
	"sshAddress":"ssh ubuntu@vm.example.test -p 24808","sshPassword":"hunter2","sshConnectionOpen":true
}}`

// The opt-in ssh command is the one place the CLI hands out the SSH password
// (`session view` and its JSON hide it): human mode prints the ready-to-run
// command and the password on two lines, nothing else.
func TestSSHCmd_HumanPrintsCommandAndPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sshSessionJSON)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newSSHCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	want := "ssh -o StrictHostKeyChecking=no -p 24808 ubuntu@vm.example.test\nhunter2\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func TestSSHCmd_PasswordOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sshSessionJSON)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newSSHCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession, "--password-only"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if stdout != "hunter2\n" {
		t.Errorf("stdout = %q, want just the password", stdout)
	}
	if _, _, err := cmdtest.Run(t, newSSHCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession, "--password-only"}, Format: output.FormatJSON}); err == nil || !strings.Contains(err.Error(), "json") {
		t.Errorf("--password-only with --format json: error = %v, want a rejection", err)
	}
}

func TestSSHCmd_JSONDecomposesAddress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sshSessionJSON)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newSSHCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got internalrde.SSHCredentials
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got.Host != "vm.example.test" || got.Port != 24808 || got.User != "ubuntu" || got.Password != "hunter2" {
		t.Errorf("credentials = %+v", got)
	}
}

// A session whose SSH endpoint is not open yet must fail the same way exec
// does, not print half a bundle.
func TestSSHCmd_NotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING","sshAddress":"ssh ubuntu@vm.example.test -p 1","sshConnectionOpen":false}}`)
	}))
	defer srv.Close()

	if _, _, err := cmdtest.Run(t, newSSHCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidSession}}); err == nil || !strings.Contains(err.Error(), "not ready") {
		t.Errorf("error = %v, want an SSH-not-ready error", err)
	}
}
