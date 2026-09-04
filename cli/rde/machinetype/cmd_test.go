package machinetype

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

func TestListCmd_RequiresStackFlag(t *testing.T) {
	_, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil || !strings.Contains(err.Error(), "stack") {
		t.Fatalf("error = %v, want it to mention required --stack flag", err)
	}
}

func TestListCmd_HidesClusterWhenUnambiguous(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1"]}]}`,
		`{"machineTypes":[
			{"id":"m-1","name":"g2.mac.m2pro.4c","clusterName":"c1","title":"M2 Pro Large","cpu":"4 vCPU","ram":"6 GB"},
			{"id":"m-2","name":"g2.mac.m1.8c","clusterName":"c1","title":"M1 Large","cpu":"8 vCPU","ram":"16 GB"}
		]}`,
	)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "osx-xcode-16.0.x-edge"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// Shows the contract name plus the backend's friendly title and specs.
	for _, want := range []string{"g2.mac.m2pro.4c", "M2 Pro Large", "4 vCPU", "6 GB"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "CLUSTER") || strings.Contains(stdout, "c1") {
		t.Errorf("unambiguous case should hide CLUSTER column, got:\n%s", stdout)
	}
}

func TestListCmd_ShowsClusterWhenAmbiguous(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1","c2"]}]}`,
		`{"machineTypes":[
			{"id":"m-1","name":"g2.mac.m2pro.4c","clusterName":"c1"},
			{"id":"m-2","name":"g2.mac.m2pro.4c","clusterName":"c2"},
			{"id":"m-3","name":"g2.mac.m1.8c","clusterName":"c1"}
		]}`,
	)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "osx-xcode-16.0.x-edge"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "CLUSTER") {
		t.Errorf("ambiguous case should show CLUSTER column, got:\n%s", stdout)
	}
	for _, want := range []string{"c1", "c2"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing cluster %q:\n%s", want, stdout)
		}
	}
}

func TestListCmd_FiltersByStackCluster(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1"]}]}`,
		`{"machineTypes":[
			{"id":"m-1","name":"g2.mac.m2pro.4c","clusterName":"c1"},
			{"id":"m-2","name":"g3.linux.8c","clusterName":"c2"}
		]}`,
	)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "osx-xcode-16.0.x-edge"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "g2.mac.m2pro.4c") {
		t.Errorf("stdout missing matching MT:\n%s", stdout)
	}
	if strings.Contains(stdout, "g3.linux.8c") {
		t.Errorf("stdout should have excluded MT from a non-matching cluster, got:\n%s", stdout)
	}
}

func TestListCmd_UnknownStack(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1"]}]}`,
		`{"machineTypes":[]}`,
	)
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "does-not-exist"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want it to mention stack not found", err)
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1"]}]}`,
		`{"machineTypes":[{"id":"m-1","name":"g2.mac","clusterName":"c1","title":"M2 Pro Large","cpu":"4 vCPU","ram":"6 GB","os":"macos"}]}`,
	)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "osx-xcode-16.0.x-edge"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Format:             output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			ClusterName string `json:"cluster_name"`
			Title       string `json:"title"`
			CPU         string `json:"cpu"`
			RAM         string `json:"ram"`
			OS          string `json:"os"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 {
		t.Fatalf("unexpected JSON items: %+v", got.Items)
	}
	it := got.Items[0]
	if it.ID != "m-1" || it.Name != "g2.mac" || it.Title != "M2 Pro Large" || it.CPU != "4 vCPU" || it.RAM != "6 GB" || it.OS != "macos" {
		t.Errorf("unexpected JSON item: %+v", it)
	}
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := catalogServer(t,
		`{"stacks":[{"id":"osx-xcode-16.0.x-edge","clusterNames":["c1"]}]}`,
		`{"machineTypes":[]}`,
	)
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{
		Args:               []string{"--stack", "osx-xcode-16.0.x-edge"},
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "No machine types available") {
		t.Errorf("expected empty-state message, got: %q", stdout)
	}
}

func TestParentCmd_ReportsMissingStackFlag(t *testing.T) {
	_, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil || !strings.Contains(err.Error(), `required flag(s) "stack" not set`) {
		t.Fatalf("error = %v, want a missing required flag error", err)
	}
}

func TestParentCmd_RejectsUnknownSubcommand(t *testing.T) {
	_, _, err := cmdtest.Run(t, NewCmd(), cmdtest.Opts{Args: []string{"lst"}, RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil || !strings.Contains(err.Error(), `unknown command "lst"`) {
		t.Fatalf("error = %v, want an unknown command error", err)
	}
}

// catalogServer returns a test server that serves the two upstream endpoints
// (/stacks and /machine-types) the command joins on.
func catalogServer(t *testing.T, stacksJSON, machineTypesJSON string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/workspaces/ws-1/stacks":
			_, _ = io.WriteString(w, stacksJSON)
		case "/v1/workspaces/ws-1/machine-types":
			_, _ = io.WriteString(w, machineTypesJSON)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}
