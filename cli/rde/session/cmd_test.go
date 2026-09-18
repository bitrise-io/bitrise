package session

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
)

func TestParentCmd_DelegatesToList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"sessions":[{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING"}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "s-1") {
		t.Errorf("bare parent should print the session list, got:\n%s", stdout)
	}
}

func TestParentCmd_RejectsUnknownSubcommand(t *testing.T) {
	// Runs the group under a parent, as in the real command tree: cobra's own
	// unknown-command check only fires on a root command, and DelegateToList
	// invokes list's RunE directly, so without Args the typo would silently
	// list sessions.
	parent := &cobra.Command{Use: "rde"}
	parent.AddCommand(NewCmd())

	_, _, err := cmdtest.Run(t, parent, cmdtest.Opts{
		Args:               []string{"session", "viwe"},
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
	})
	if err == nil || !strings.Contains(err.Error(), `unknown command "viwe"`) {
		t.Fatalf("error = %v, want an unknown command error", err)
	}
}
