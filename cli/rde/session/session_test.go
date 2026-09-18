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

func TestListCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces/ws-1/sessions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("auth = %q", got)
		}
		_, _ = io.WriteString(w, `{"sessions":[
			{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING","templateSnapshot":{"templateName":"tmpl"}}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"dev", "running", "tmpl", "s-1"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"sessions":[{"id":"s-1","name":"dev","status":"SESSION_STATUS_RUNNING"}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 || got.Items[0].ID != "s-1" || got.Items[0].Status != "running" {
		t.Errorf("unexpected JSON: %+v", got.Items)
	}
}

func TestListCmd_LabelSelectorsQuery(t *testing.T) {
	var gotSelectors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSelectors = r.URL.Query()["labelSelectors"]
		_, _ = io.WriteString(w, `{"sessions":[]}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"-l", "team=mobile", "--label-selector", "branch=main"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(gotSelectors) != 2 || gotSelectors[0] != "team=mobile" || gotSelectors[1] != "branch=main" {
		t.Errorf("labelSelectors = %v, want [team=mobile branch=main]", gotSelectors)
	}
}

func TestListCmd_ScopeQuery(t *testing.T) {
	var gotScope string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotScope = r.URL.Query().Get("scope")
		_, _ = io.WriteString(w, `{"sessions":[]}`)
	}))
	defer srv.Close()

	// Default scope is mine, sent as the backend's enum name.
	if _, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotScope != "SESSION_LIST_SCOPE_MINE" {
		t.Errorf("default scope = %q, want SESSION_LIST_SCOPE_MINE", gotScope)
	}

	if _, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--scope", "workspace"},
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotScope != "SESSION_LIST_SCOPE_WORKSPACE" {
		t.Errorf("scope = %q, want SESSION_LIST_SCOPE_WORKSPACE", gotScope)
	}
}

func TestListCmd_InvalidScopeErrors(t *testing.T) {
	// Validation happens before any HTTP call.
	_, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--scope", "everyone"},
	})
	if err == nil || !strings.Contains(err.Error(), "--scope") {
		t.Errorf("error = %v, want scope validation error", err)
	}
}

func TestListCmd_MalformedSelectorErrors(t *testing.T) {
	_, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"-l", "bare-key"},
	})
	if err == nil || !strings.Contains(err.Error(), "--label-selector") {
		t.Errorf("error = %v, want selector parse error", err)
	}
}
