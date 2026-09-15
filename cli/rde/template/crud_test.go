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

// TestCreateCmd_DeviceFlags: the --device-* flags declare the template's
// device and go out as deviceSpec (taking precedence over the file).
func TestCreateCmd_DeviceFlags(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-new","name":"iOS","deviceSpec":{"platform":"ios","deviceModel":"iPhone 16"}}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--file", "-", "--device-platform", "ios", "--device-model", "iPhone 16", "--device-os-version", "18.2"},
		Stdin:              `{"name":"iOS","stack_id":"osx-xcode-16.0.x-edge","machine_type":"g2.mac","device_spec":{"platform":"android"}}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	spec, _ := gotBody["deviceSpec"].(map[string]any)
	if spec["platform"] != "ios" || spec["deviceModel"] != "iPhone 16" || spec["osVersion"] != "18.2" {
		t.Errorf("unexpected deviceSpec: %v (body=%v)", gotBody["deviceSpec"], gotBody)
	}
	if !strings.Contains(stdout, "iOS simulator · iPhone 16") {
		t.Errorf("stdout missing the declared device:\n%s", stdout)
	}
}

// TestCreateCmd_DeviceSpecFromFile: device_spec in the spec file is
// forwarded when no flag overrides it.
func TestCreateCmd_DeviceSpecFromFile(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-new","name":"Android"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--file", "-"},
		Stdin:              `{"name":"Android","stack_id":"ubuntu-android","machine_type":"standard","device_spec":{"platform":"android","system_image":"system-images;android-34;google_apis;x86_64"}}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	spec, _ := gotBody["deviceSpec"].(map[string]any)
	if spec["platform"] != "android" || spec["systemImage"] != "system-images;android-34;google_apis;x86_64" {
		t.Errorf("unexpected deviceSpec: %v", gotBody["deviceSpec"])
	}
}

func TestCreateCmd_DeviceFlagsValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be hit when the device flags are invalid")
	}))
	defer srv.Close()

	for name, args := range map[string][]string{
		"model needs platform":   {"--file", "-", "--device-model", "iPhone 16"},
		"bad platform":           {"--file", "-", "--device-platform", "windows"},
		"system image on iOS":    {"--file", "-", "--device-platform", "ios", "--device-system-image", "system-images;android-34;google_apis;x86_64"},
		"os version no platform": {"--file", "-", "--device-os-version", "18.2"},
	} {
		_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
			RDEAPIBaseURL:      srv.URL,
			DefaultWorkspaceID: "ws-1",
			Args:               args,
			Stdin:              `{"name":"Dev","stack_id":"s","machine_type":"m"}`,
		})
		if err == nil || !strings.Contains(err.Error(), "--device-") {
			t.Errorf("%s: error = %v, want a --device-* validation error", name, err)
		}
	}
}

// TestUpdateCmd_DeviceFlags: any --device-* flag replaces the template's
// device — deviceSpec plus updateDeviceSpec=true — and works without --file.
func TestUpdateCmd_DeviceFlags(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workspaces/ws-1/templates/"+uuidTemplateID {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-1","name":"Android"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidTemplateID, "--device-platform", "android", "--device-model", "pixel_7"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	spec, _ := gotBody["deviceSpec"].(map[string]any)
	if spec["platform"] != "android" || spec["deviceModel"] != "pixel_7" {
		t.Errorf("unexpected deviceSpec: %v", gotBody["deviceSpec"])
	}
	if gotBody["updateDeviceSpec"] != true {
		t.Errorf("updateDeviceSpec = %v, want true (body=%v)", gotBody["updateDeviceSpec"], gotBody)
	}
	// Nothing else may be touched by a device-only update.
	for _, k := range []string{"name", "stackId", "updateSessionInputs"} {
		if _, ok := gotBody[k]; ok {
			t.Errorf("%s should be absent, body=%v", k, gotBody)
		}
	}
}

// TestUpdateCmd_ClearDevice: --clear-device sends updateDeviceSpec=true with
// no deviceSpec, and refuses to combine with a --device-* flag.
func TestUpdateCmd_ClearDevice(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-1","name":"Dev"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidTemplateID, "--clear-device"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["updateDeviceSpec"] != true {
		t.Errorf("updateDeviceSpec = %v, want true (body=%v)", gotBody["updateDeviceSpec"], gotBody)
	}
	if _, ok := gotBody["deviceSpec"]; ok {
		t.Errorf("deviceSpec must be absent when clearing, body=%v", gotBody)
	}

	gotBody = nil
	_, _, err = cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidTemplateID, "--clear-device", "--device-platform", "ios"},
	})
	if err == nil || !strings.Contains(err.Error(), "clear-device") {
		t.Fatalf("expected a --clear-device exclusivity error, got %v", err)
	}
	if gotBody != nil {
		t.Errorf("server must not be hit when the flags conflict")
	}
}

// TestUpdateCmd_FileWithoutDeviceLeavesItAlone: a spec file with no
// device_spec must not touch the template's device.
func TestUpdateCmd_FileWithoutDeviceLeavesItAlone(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"template":{"id":"t-1","name":"Renamed"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidTemplateID, "--file", "-"},
		Stdin:              `{"name":"Renamed"}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, ok := gotBody["updateDeviceSpec"]; ok {
		t.Errorf("updateDeviceSpec should be absent, body=%v", gotBody)
	}
}
