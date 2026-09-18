package template

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
)

// uuidTemplateID is a UUID-shaped template arg. Real RDE template IDs are
// UUIDs, so passing one exercises the ResolveTemplateID short-circuit (no
// extra ListTemplates call) — the path production hits when a user pastes an
// ID rather than a name.
const uuidTemplateID = "33333333-4444-4444-8444-555555555555"

func TestCreateCmd_HappyPath(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws-1/templates" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-new","name":"Dev","stackId":"linux-ubuntu-24.04","machineType":"standard"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--file", "-"},
		Stdin:              `{"name":"Dev","stack_id":"linux-ubuntu-24.04","machine_type":"standard","session_inputs":[{"key":"repo","required":true}]}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// The spec's snake_case fields must map onto the camelCase wire body.
	if gotBody["name"] != "Dev" || gotBody["stackId"] != "linux-ubuntu-24.04" || gotBody["machineType"] != "standard" {
		t.Errorf("unexpected create body: %v", gotBody)
	}
	if !strings.Contains(stdout, "t-new") {
		t.Errorf("stdout missing new template ID:\n%s", stdout)
	}
}

func TestCreateCmd_RequiresFile(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil || !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %v, want --file required", err)
	}
}

func TestCreateCmd_MissingRequiredSpecField(t *testing.T) {
	// machine_type is required by the service; the spec omits it, so the
	// command must fail before any HTTP call.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be hit when the spec is invalid")
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--file", "-"},
		Stdin:              `{"name":"Dev","stack_id":"linux-ubuntu-24.04"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "machine_type") {
		t.Errorf("error = %v, want machine_type-required", err)
	}
}

func TestCreateCmd_MalformedJSON(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--file", "-"},
		Stdin:              `{not json`,
	})
	if err == nil || !strings.Contains(err.Error(), "parse template spec") {
		t.Errorf("error = %v, want parse error", err)
	}
}

func TestUpdateCmd_SendsReplaceFlagForArrays(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workspaces/ws-1/templates/"+uuidTemplateID {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-1","name":"Renamed"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidTemplateID, "--file", "-"},
		Stdin:              `{"name":"Renamed","session_inputs":[{"key":"repo"}]}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["name"] != "Renamed" {
		t.Errorf("name = %v, want Renamed", gotBody["name"])
	}
	// A present array field must carry its updateXxx replace flag so the
	// server replaces (not merges) the list.
	if gotBody["updateSessionInputs"] != true {
		t.Errorf("updateSessionInputs = %v, want true (body=%v)", gotBody["updateSessionInputs"], gotBody)
	}
	// An absent array field must not trigger its flag.
	if _, ok := gotBody["updateFeatureFlags"]; ok {
		t.Errorf("updateFeatureFlags should be absent, body=%v", gotBody)
	}
}

func TestUpdateCmd_RequiresFile(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1", Args: []string{"t-1"}})
	if err == nil || !strings.Contains(err.Error(), "--file") {
		t.Errorf("error = %v, want --file required", err)
	}
}

func TestUpdateCmd_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1", Args: []string{"--file", "-"}})
	if err == nil {
		t.Fatal("expected error when TEMPLATE_ID is missing")
	}
}

func TestDeleteCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/workspaces/ws-1/templates/"+uuidTemplateID {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{uuidTemplateID}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// Confirmation goes to stderr, never stdout.
	if !strings.Contains(stderr, "Deleted template "+uuidTemplateID) {
		t.Errorf("stderr missing confirmation: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout should be empty for delete, got: %q", stdout)
	}
}

// TestDeleteCmd_ResolvesName covers name → ID resolution: a non-UUID arg is
// treated as a template name, looked up via ListTemplates, and the resolved ID
// is what gets deleted.
func TestDeleteCmd_ResolvesName(t *testing.T) {
	var deletedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/templates":
			_, _ = io.WriteString(w, `{"templates":[{"id":"t-9","name":"Linux Dev"},{"id":"t-7","name":"macOS Dev"}]}`)
		case r.Method == http.MethodDelete:
			deletedPath = r.URL.Path
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	_, stderr, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{RDEAPIBaseURL: srv.URL, DefaultWorkspaceID: "ws-1", Args: []string{"Linux Dev"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if deletedPath != "/v1/workspaces/ws-1/templates/t-9" {
		t.Errorf("deleted path = %q, want the resolved id t-9", deletedPath)
	}
	if !strings.Contains(stderr, "Deleted template t-9") {
		t.Errorf("stderr should confirm deletion of the resolved id: %q", stderr)
	}
}

func TestDeleteCmd_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{RDEAPIBaseURL: "http://unused", DefaultWorkspaceID: "ws-1"})
	if err == nil {
		t.Fatal("expected error when TEMPLATE_ID is missing")
	}
}
