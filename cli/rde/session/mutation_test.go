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

func TestCreateCmd_HappyPath(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws-1/sessions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--input", "repo=my-app"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["templateId"] != uuidTemplate || gotBody["name"] != "dev" {
		t.Errorf("unexpected create body: %v", gotBody)
	}
	if !strings.Contains(stdout, "Session created") || !strings.Contains(stdout, "s-new") {
		t.Errorf("stdout missing create confirmation:\n%s", stdout)
	}
}

func TestCreateCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"},
			"autoMappedInputs":[{"sessionInputKey":"gh","savedInputId":"sv-1"}]}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--map-saved-inputs"},
		Format:             output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Session struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"session"`
		AutoMapped []struct {
			SessionInputKey string `json:"session_input_key"`
		} `json:"auto_mapped_inputs"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got.Session.ID != "s-new" || got.Session.Status != "pending" {
		t.Errorf("unexpected session JSON: %+v", got.Session)
	}
	if len(got.AutoMapped) != 1 || got.AutoMapped[0].SessionInputKey != "gh" {
		t.Errorf("unexpected auto-mapped JSON: %+v", got.AutoMapped)
	}
}

func TestCreateCmd_RequiresName(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--template", uuidTemplate},
	})
	if err == nil || !strings.Contains(err.Error(), "NAME") {
		t.Errorf("error = %v, want NAME positional required", err)
	}
}

func TestCreateCmd_AutoTerminateZeroIsSent(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev"}}`)
	}))
	defer srv.Close()

	// Explicitly setting --auto-terminate-minutes 0 must send 0 (disable),
	// distinct from omitting the flag (backend default).
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--auto-terminate-minutes", "0"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	v, ok := gotBody["autoTerminateMinutes"]
	if !ok {
		t.Fatalf("autoTerminateMinutes should be present when flag is set, body=%v", gotBody)
	}
	if v != float64(0) {
		t.Errorf("autoTerminateMinutes = %v, want 0", v)
	}
}

func TestCreateCmd_AutoTerminateOmittedWhenFlagUnset(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, ok := gotBody["autoTerminateMinutes"]; ok {
		t.Errorf("autoTerminateMinutes should be omitted when flag unset, body=%v", gotBody)
	}
}

func TestCreateCmd_TemplateLess(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws-1/sessions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--stack", "osx-xcode-16.0.x-edge", "--machine-type", "g2.mac.m2pro.6c-14g"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, ok := gotBody["templateId"]; ok {
		t.Errorf("templateId should be omitted for a template-less session, body=%v", gotBody)
	}
	if gotBody["stackId"] != "osx-xcode-16.0.x-edge" || gotBody["machineType"] != "g2.mac.m2pro.6c-14g" {
		t.Errorf("unexpected create body: %v", gotBody)
	}
	if !strings.Contains(stdout, "Session created") || !strings.Contains(stdout, "s-new") {
		t.Errorf("stdout missing create confirmation:\n%s", stdout)
	}
}

func TestCreateCmd_StackOverridesWithTemplate(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev"}}`)
	}))
	defer srv.Close()

	// --stack / --machine-type may accompany a template to override its
	// defaults; the template ID is still sent.
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--stack", "osx-xcode-16.0.x-edge", "--machine-type", "g2.mac.m2pro.6c-14g"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["templateId"] != uuidTemplate {
		t.Errorf("templateId = %v, want %s", gotBody["templateId"], uuidTemplate)
	}
	if gotBody["stackId"] != "osx-xcode-16.0.x-edge" || gotBody["machineType"] != "g2.mac.m2pro.6c-14g" {
		t.Errorf("unexpected create body: %v", gotBody)
	}
}

func TestCreateCmd_RequiresTemplateOrMachineSpec(t *testing.T) {
	cases := map[string][]string{
		"no template, no machine spec": {"dev"},
		"stack without machine type":   {"dev", "--stack", "osx-xcode-16.0.x-edge"},
		"machine type without stack":   {"dev", "--machine-type", "g2.mac.m2pro.6c-14g"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
				RDEAPIBaseURL:      "http://unused",
				DefaultWorkspaceID: "ws-1",
				Args:               args,
			})
			if err == nil || !strings.Contains(err.Error(), "--machine-type") {
				t.Errorf("error = %v, want template-or-stack/machine-type requirement", err)
			}
		})
	}
}

func TestCreateCmd_LabelsSent(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "-l", "branch=main", "--label", "team=mobile"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	labels, ok := gotBody["labels"].(map[string]any)
	if !ok || labels["branch"] != "main" || labels["team"] != "mobile" {
		t.Errorf("labels = %v, want branch=main team=mobile", gotBody["labels"])
	}
}

func TestCreateCmd_LabelsOmittedWhenUnset(t *testing.T) {
	// No --label — the body must not carry a labels field at all.
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, ok := gotBody["labels"]; ok {
		t.Errorf("labels should be omitted, body=%v", gotBody)
	}
}

func TestCreateCmd_MalformedLabelErrors(t *testing.T) {
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--label", "no-equals"},
	})
	if err == nil || !strings.Contains(err.Error(), "--label") {
		t.Errorf("error = %v, want --label parse error", err)
	}
}

func TestCreateCmd_WaitBlocksUntilRunning(t *testing.T) {
	var getCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions":
			_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/s-new":
			getCount++
			// Report the session as running so WaitForReady settles on the
			// first poll and no poll interval elapses.
			_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_RUNNING"}}`)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--wait"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if getCount < 1 {
		t.Errorf("expected --wait to poll GetSession after create, got %d GETs", getCount)
	}
	if !strings.Contains(stdout, "running") || !strings.Contains(stdout, "s-new") {
		t.Errorf("stdout should show the final running status:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Waiting for session") {
		t.Errorf("stderr should announce the wait:\n%s", stderr)
	}
}

func TestCreateCmd_WaitNonRunningExitsNonZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions":
			_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/s-new":
			// Provisioning failed — --wait must surface it as a non-zero exit.
			_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_FAILED"}}`)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--wait"},
	})
	if err == nil || !strings.Contains(err.Error(), "expected running") {
		t.Errorf("error = %v, want a non-running final-status error", err)
	}
	if !strings.Contains(stdout, "s-new") {
		t.Errorf("stdout should still carry the created session:\n%s", stdout)
	}
}

func TestCreateCmd_WaitErrorStillRendersSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions":
			_, _ = io.WriteString(w, `{"session":{"id":"s-new","name":"dev","status":"SESSION_STATUS_PENDING"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/s-new":
			// A --wait-timeout expiry is the real-world trigger; a failing
			// poll reaches the same path without depending on timing.
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev", "--template", uuidTemplate, "--wait"},
		Format:             output.FormatJSON,
	})
	if err == nil || !strings.Contains(err.Error(), "waiting for session") {
		t.Errorf("error = %v, want a wait error", err)
	}
	// The session was created and is billing: its ID must survive the failed
	// wait, including in a machine-readable format where the stderr
	// breadcrumb carrying it is suppressed.
	if !strings.Contains(stdout, "s-new") {
		t.Errorf("stdout should carry the created session despite the wait error:\n%s", stdout)
	}
}

func TestUpdateCmd_RequiresAField(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"s-1"},
	})
	if err == nil || !strings.Contains(err.Error(), "--name") {
		t.Errorf("error = %v, want at-least-one-field error", err)
	}
}

func TestUpdateCmd_OnlySetFieldsSent(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workspaces/ws-1/sessions/"+uuidSession {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"renamed"}}`)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--name", "renamed"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotBody["name"] != "renamed" {
		t.Errorf("name = %v, want renamed", gotBody["name"])
	}
	// --description / --auto-terminate-minutes weren't set, so they drop out.
	if _, ok := gotBody["description"]; ok {
		t.Errorf("description should be omitted, body=%v", gotBody)
	}
	if _, ok := gotBody["autoTerminateMinutes"]; ok {
		t.Errorf("autoTerminateMinutes should be omitted, body=%v", gotBody)
	}
}

func TestUpdateCmd_LabelsSentAndRendered(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = io.WriteString(w, `{"session":{"id":"`+uuidSession+`","name":"dev","status":"SESSION_STATUS_RUNNING","labels":{"branch":"main","team":"mobile"}}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--label", "branch=main", "--unset-label", "wip"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	labels, ok := gotBody["labels"].(map[string]any)
	if !ok || labels["branch"] != "main" {
		t.Errorf("labels = %v, want branch=main", gotBody["labels"])
	}
	removed, ok := gotBody["removeLabels"].([]any)
	if !ok || len(removed) != 1 || removed[0] != "wip" {
		t.Errorf("removeLabels = %v, want [wip]", gotBody["removeLabels"])
	}
	// The human detail view renders the returned labels.
	for _, want := range []string{"Labels", "branch", "main", "team", "mobile"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestUpdateCmd_LabelSetAndUnsetConflict(t *testing.T) {
	_, _, err := cmdtest.Run(t, newUpdateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--label", "wip=yes", "--unset-label", "wip"},
	})
	if err == nil || !strings.Contains(err.Error(), "both set") {
		t.Errorf("error = %v, want set/unset conflict error", err)
	}
}

func TestTerminateCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws-1/sessions/"+uuidSession+"/terminate" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_TERMINATING"}}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newTerminateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "terminating") {
		t.Errorf("stdout missing status:\n%s", stdout)
	}
}

func TestRestoreCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession:
			// Pre-flight disk-status check before the restore call.
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_TERMINATED","persistentDiskStatus":"PERSISTENT_DISK_STATUS_AVAILABLE"}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession+"/restore":
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_STARTING"}}`)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newRestoreCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "starting") {
		t.Errorf("stdout missing status:\n%s", stdout)
	}
}

func TestRestoreCmd_DiskUnavailableFailsFast(t *testing.T) {
	var restoreCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession:
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_TERMINATED","persistentDiskStatus":"PERSISTENT_DISK_STATUS_UNAVAILABLE"}}`)
		case r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession+"/restore":
			restoreCalled = true
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newRestoreCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession},
	})
	if err == nil {
		t.Fatal("expected restore to fail when the persistent disk is unavailable")
	}
	if !strings.Contains(err.Error(), "no longer available") {
		t.Errorf("error missing reason: %v", err)
	}
	if restoreCalled {
		t.Error("restore endpoint should not be called when the disk is unavailable")
	}
}

func TestRestoreCmd_DiskExpiringSoonWarnsButProceeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession:
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_TERMINATED","persistentDiskStatus":"PERSISTENT_DISK_STATUS_UNAVAILABLE_SOON"}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession+"/restore":
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_STARTING"}}`)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newRestoreCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stderr, "unavailable soon") {
		t.Errorf("stderr missing expiry warning:\n%s", stderr)
	}
	if !strings.Contains(stdout, "starting") {
		t.Errorf("stdout missing status:\n%s", stdout)
	}
}

func TestRestoreCmd_WaitBlocksUntilRunning(t *testing.T) {
	// Keep the id consistent across pre-flight GET, restore POST, and the
	// wait poll (WaitForReady polls the id from the restore response body).
	sess := func(status string) string {
		return `{"session":{"id":"` + uuidSession + `","name":"dev","status":"` + status + `"}}`
	}
	var getCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession:
			getCount++
			if getCount == 1 {
				// Pre-flight disk-status check before the restore call.
				_, _ = io.WriteString(w, `{"session":{"id":"`+uuidSession+`","name":"dev","status":"SESSION_STATUS_TERMINATED","persistentDiskStatus":"PERSISTENT_DISK_STATUS_AVAILABLE"}}`)
				return
			}
			// WaitForReady poll: report the session as running so it settles
			// immediately (no poll interval elapses).
			_, _ = io.WriteString(w, sess("SESSION_STATUS_RUNNING"))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession+"/restore":
			_, _ = io.WriteString(w, sess("SESSION_STATUS_STARTING"))
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newRestoreCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--wait"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if getCount < 2 {
		t.Errorf("expected --wait to poll GetSession after restore, got %d GETs", getCount)
	}
	if !strings.Contains(stdout, "running") {
		t.Errorf("stdout should show the final running status:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Waiting for session") {
		t.Errorf("stderr should announce the wait:\n%s", stderr)
	}
}

func TestRestoreCmd_WaitNonRunningExitsNonZero(t *testing.T) {
	sess := func(status string) string {
		return `{"session":{"id":"` + uuidSession + `","name":"dev","status":"` + status + `"}}`
	}
	var getCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession:
			getCount++
			if getCount == 1 {
				_, _ = io.WriteString(w, `{"session":{"id":"`+uuidSession+`","name":"dev","status":"SESSION_STATUS_TERMINATED","persistentDiskStatus":"PERSISTENT_DISK_STATUS_AVAILABLE"}}`)
				return
			}
			// Provisioning failed — --wait must surface it as a non-zero exit.
			_, _ = io.WriteString(w, sess("SESSION_STATUS_FAILED"))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession+"/restore":
			_, _ = io.WriteString(w, sess("SESSION_STATUS_STARTING"))
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newRestoreCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--wait"},
	})
	if err == nil || !strings.Contains(err.Error(), "expected running") {
		t.Errorf("error = %v, want a non-running final-status error", err)
	}
}

func TestDeleteCmd_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/workspaces/ws-1/sessions/"+uuidSession {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	stdout, stderr, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stderr, "Deleted session "+uuidSession) {
		t.Errorf("stderr missing confirmation: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout should be empty for delete, got: %q", stdout)
	}
}

// TestDeleteCmd_ResolvesName: a non-UUID arg is treated as a session name,
// looked up via ListSessions, and the resolved ID is what actually gets
// deleted.
func TestDeleteCmd_ResolvesName(t *testing.T) {
	var deletedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions":
			_, _ = io.WriteString(w, `{"sessions":[{"id":"s-9","name":"my-box"},{"id":"s-7","name":"other"}]}`)
		case r.Method == http.MethodDelete:
			deletedPath = r.URL.Path
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	_, stderr, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"my-box"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if deletedPath != "/v1/workspaces/ws-1/sessions/s-9" {
		t.Errorf("deleted path = %q, want the resolved id s-9", deletedPath)
	}
	if !strings.Contains(stderr, "Deleted session s-9") {
		t.Errorf("stderr should confirm deletion of the resolved id: %q", stderr)
	}
}

// TestDeleteCmd_AmbiguousNameError pins the non-uniqueness guard: when a name
// matches more than one session, delete refuses rather than guessing.
func TestDeleteCmd_AmbiguousNameError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions" {
			_, _ = io.WriteString(w, `{"sessions":[{"id":"s-1","name":"dup"},{"id":"s-2","name":"dup"}]}`)
			return
		}
		t.Errorf("nothing should be deleted for an ambiguous name: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	_, _, err := cmdtest.Run(t, newDeleteCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dup"},
	})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error = %v, want ambiguous-name error", err)
	}
}

func TestDeleteTerminatedCmd_YesSkipsPrompt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws-1/sessions:delete-terminated" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"deletedCount":3}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newDeleteTerminatedCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--yes"},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "Deleted 3 terminated session(s)") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}

func TestDeleteTerminatedCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"deletedCount":3}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newDeleteTerminatedCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"--yes"},
		Format:             output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		DeletedCount int `json:"deleted_count"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal JSON output: %v\n%s", err, stdout)
	}
	if got.DeletedCount != 3 {
		t.Errorf("deleted_count = %d, want 3", got.DeletedCount)
	}
}

func TestDeleteTerminatedCmd_AbortsOnNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be hit when the user declines confirmation")
	}))
	defer srv.Close()

	// No --yes; feed "n" to the confirmation prompt.
	_, _, err := cmdtest.Run(t, newDeleteTerminatedCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Stdin:              "n\n",
	})
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error = %v, want aborted", err)
	}
}

func TestDeleteTerminatedCmd_ProceedsOnYes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"deletedCount":1}`)
	}))
	defer srv.Close()

	stdout, _, err := cmdtest.Run(t, newDeleteTerminatedCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Stdin:              "y\n",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "Deleted 1 terminated session(s)") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}
