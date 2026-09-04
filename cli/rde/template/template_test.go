package template

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
		if r.URL.Path != "/v1/workspaces/ws-1/templates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"templates":[
			{"id":"t-1","name":"Linux Dev","stackId":"linux-ubuntu-24.04","machineType":"standard","createdByEmail":"a@b.io"}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"Linux Dev", "linux-ubuntu-24.04", "standard", "a@b.io", "t-1"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"templates":[
			{"id":"t-1","name":"Linux Dev","stackId":"linux-ubuntu-24.04","machineType":"standard"}
		]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			MachineType string `json:"machine_type"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if len(got.Items) != 1 || got.Items[0].ID != "t-1" || got.Items[0].MachineType != "standard" {
		t.Errorf("unexpected JSON items: %+v", got.Items)
	}
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"templates":[]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newListCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "No templates found.") {
		t.Errorf("expected empty-state message, got: %q", stdout)
	}
}

func TestViewCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces/ws-1/templates/"+uuidTemplateID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"template":{
			"id":"t-1","name":"Linux Dev","stackId":"linux-ubuntu-24.04","machineType":"standard",
			"sessionInputs":[{"key":"repo","required":true,"description":"Repo to clone"}],
			"templateVariables":[{"key":"TOKEN","isSecret":true}]
		}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidTemplateID}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, want := range []string{"Linux Dev", "t-1", "linux-ubuntu-24.04", "repo", "(required)", "TOKEN", "(secret)"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestViewCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"template":{"id":"t-1","name":"Linux Dev","machineType":"standard"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidTemplateID}, Format: output.FormatJSON})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got["id"] != "t-1" || got["machine_type"] != "standard" {
		t.Errorf("unexpected JSON: %v", got)
	}
}

func TestViewCmd_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, newViewCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil {
		t.Fatal("expected error when TEMPLATE_ID is missing")
	}
}
