package stack

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
		if r.URL.Path != "/v1/workspaces/ws-1/stacks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"stacks":[{"id":"osx-xcode-16.0.x-edge","title":"Xcode 16.0","os":"macos","osVersion":26,"status":"edge","clusterNames":["c1"]}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"osx-xcode-16.0.x-edge", "Xcode 16.0", "macOS 26", "edge"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "c1") {
		t.Errorf("human output should not expose cluster, got:\n%s", stdout)
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"stacks":[{"id":"osx-xcode-16.0.x-edge","title":"Xcode 16.0","os":"macos","osVersion":26,"status":"edge","clusterNames":["c1"]}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			OS        string `json:"os"`
			OSVersion int32  `json:"os_version"`
			Status    string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 {
		t.Fatalf("unexpected JSON items: %+v", got.Items)
	}
	it := got.Items[0]
	if it.ID != "osx-xcode-16.0.x-edge" || it.Title != "Xcode 16.0" || it.OS != "macos" || it.OSVersion != 26 || it.Status != "edge" {
		t.Errorf("unexpected JSON item: %+v", it)
	}
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"stacks":[]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "No stacks found.") {
		t.Errorf("expected empty-state message, got: %q", stdout)
	}
}
