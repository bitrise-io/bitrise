package rdeapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListSessions_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"sessions":[
		{"id":"s1","name":"dev","status":"SESSION_STATUS_RUNNING","templateSnapshot":{"templateName":"tmpl","stackId":"osx-xcode-16.0.x-edge"},"labels":{"team":"mobile"},"ownerType":"workspace","ownerId":"my-ws"},
		{"id":"s2","name":"old","status":"SESSION_STATUS_TERMINATED"}
	]}`)

	sessions, err := rs.client().ListSessions(context.Background(), "ws-1", nil, "")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if rs.lastMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if rs.lastQuery != "" {
		t.Errorf("query = %q, want none without selectors", rs.lastQuery)
	}
	if len(sessions) != 2 {
		t.Fatalf("got %d sessions, want 2", len(sessions))
	}
	if sessions[0].TemplateSnapshot == nil || sessions[0].TemplateSnapshot.TemplateName != "tmpl" {
		t.Errorf("session[0] template snapshot = %+v", sessions[0].TemplateSnapshot)
	}
	// Wire format keeps the raw enum string — mapping to snake_case lives in internal/rde.
	if sessions[0].Status != "SESSION_STATUS_RUNNING" {
		t.Errorf("status = %q, want raw enum", sessions[0].Status)
	}
	if sessions[0].Labels["team"] != "mobile" {
		t.Errorf("labels = %v, want team=mobile", sessions[0].Labels)
	}
	if sessions[0].OwnerType != "workspace" || sessions[0].OwnerID != "my-ws" {
		t.Errorf("owner = %q/%q, want workspace/my-ws", sessions[0].OwnerType, sessions[0].OwnerID)
	}
}

func TestListSessions_LabelSelectorsQuery(t *testing.T) {
	rs := newRecordingServer(t, `{"sessions":[]}`)

	if _, err := rs.client().ListSessions(context.Background(), "ws-1", []string{"team=mobile", "branch=main"}, ""); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	// One repeated labelSelectors param per selector, "=" percent-encoded,
	// order preserved (grpc-gateway maps them onto the repeated proto field).
	if want := "labelSelectors=team%3Dmobile&labelSelectors=branch%3Dmain"; rs.lastQuery != want {
		t.Errorf("query = %s, want %s", rs.lastQuery, want)
	}
}

func TestListSessions_ScopeQuery(t *testing.T) {
	rs := newRecordingServer(t, `{"sessions":[]}`)

	// Scope is translated to the backend's enum name.
	if _, err := rs.client().ListSessions(context.Background(), "ws-1", nil, "workspace"); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if want := "scope=SESSION_LIST_SCOPE_WORKSPACE"; rs.lastQuery != want {
		t.Errorf("query = %q, want %q", rs.lastQuery, want)
	}

	// Scope combines with label selectors (Encode sorts by key).
	if _, err := rs.client().ListSessions(context.Background(), "ws-1", []string{"team=mobile"}, "mine"); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if want := "labelSelectors=team%3Dmobile&scope=SESSION_LIST_SCOPE_MINE"; rs.lastQuery != want {
		t.Errorf("query = %q, want %q", rs.lastQuery, want)
	}

	// Unknown scopes are omitted so the backend default applies.
	if _, err := rs.client().ListSessions(context.Background(), "ws-1", nil, "everyone"); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if rs.lastQuery != "" {
		t.Errorf("query = %q, want none for unknown scope", rs.lastQuery)
	}
}

func TestGetSession_Path(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"s1","name":"dev"}}`)

	sess, err := rs.client().GetSession(context.Background(), "ws-1", "s1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if sess.ID != "s1" {
		t.Errorf("id = %q, want s1", sess.ID)
	}
}

func TestGetSession_EscapesSessionID(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"a b"}}`)

	if _, err := rs.client().GetSession(context.Background(), "ws-1", "a b"); err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	// The raw request URI must carry the escaped form.
	if want := "/v1/workspaces/ws-1/sessions/a%20b"; rs.lastURI != want {
		t.Errorf("escaped URI = %s, want %s", rs.lastURI, want)
	}
}

func TestCreateSession_BodyAndResponse(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"new","name":"dev","status":"SESSION_STATUS_PENDING"},
		"autoMappedInputs":[{"sessionInputKey":"gh","savedInputId":"sv-1"}]}`)

	mins := 30
	sess, mapped, err := rs.client().CreateSession(context.Background(), "ws-1", CreateSessionRequest{
		Name:                    "dev",
		TemplateID:              "tmpl-1",
		SessionInputs:           []SessionInputValue{{Key: "repo", Value: "app"}},
		AutoTerminateMinutes:    &mins,
		MapSavedToSessionInputs: true,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}

	var sent CreateSessionRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent.TemplateID != "tmpl-1" {
		t.Errorf("sent templateId = %q, want tmpl-1", sent.TemplateID)
	}
	if sent.AutoTerminateMinutes == nil || *sent.AutoTerminateMinutes != 30 {
		t.Errorf("sent autoTerminateMinutes = %v, want pointer to 30", sent.AutoTerminateMinutes)
	}
	if !sent.MapSavedToSessionInputs {
		t.Error("sent mapSavedToSessionInputs = false, want true")
	}

	if sess.ID != "new" {
		t.Errorf("session id = %q, want new", sess.ID)
	}
	if len(mapped) != 1 || mapped[0].SessionInputKey != "gh" || mapped[0].SavedInputID != "sv-1" {
		t.Errorf("auto-mapped = %+v", mapped)
	}
}

func TestCreateSession_LabelsSent(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"new","name":"dev"}}`)

	if _, _, err := rs.client().CreateSession(context.Background(), "ws-1", CreateSessionRequest{
		Name:   "dev",
		Labels: map[string]string{"team": "mobile", "branch": "main"},
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	labels, ok := sent["labels"].(map[string]any)
	if !ok {
		t.Fatalf("labels missing from body: %s", rs.lastBody)
	}
	if labels["team"] != "mobile" || labels["branch"] != "main" {
		t.Errorf("sent labels = %v", labels)
	}
}

func TestCreateSession_TemplatelessOmitsTemplateID(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"new","name":"dev"}}`)

	if _, _, err := rs.client().CreateSession(context.Background(), "ws-1", CreateSessionRequest{
		Name:        "dev",
		StackID:     "linux-ubuntu-24.04",
		MachineType: "g2.linux.amd-zen5.8c-32g",
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if _, ok := sent["templateId"]; ok {
		t.Errorf("templateId should be omitted, body = %s", rs.lastBody)
	}
	if sent["stackId"] != "linux-ubuntu-24.04" {
		t.Errorf("stackId = %v, want linux-ubuntu-24.04", sent["stackId"])
	}
	if sent["machineType"] != "g2.linux.amd-zen5.8c-32g" {
		t.Errorf("machineType = %v, want g2.linux.amd-zen5.8c-32g", sent["machineType"])
	}
}

func TestUpdateSession_OmitsUnsetPointerFields(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"s1","name":"renamed"}}`)

	name := "renamed"
	if _, err := rs.client().UpdateSession(context.Background(), "ws-1", "s1", UpdateSessionRequest{Name: &name}); err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}
	if rs.lastMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", rs.lastMethod)
	}

	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["name"] != "renamed" {
		t.Errorf("name = %v, want renamed", sent["name"])
	}
	for _, field := range []string{"description", "autoTerminateMinutes", "labels", "removeLabels"} {
		if _, ok := sent[field]; ok {
			t.Errorf("%s should be omitted, body = %s", field, rs.lastBody)
		}
	}
}

func TestUpdateSession_LabelsAndRemovals(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"s1","name":"dev","labels":{"branch":"main"}}}`)

	sess, err := rs.client().UpdateSession(context.Background(), "ws-1", "s1", UpdateSessionRequest{
		Labels:       map[string]string{"branch": "main"},
		RemoveLabels: []string{"wip"},
	})
	if err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	labels, ok := sent["labels"].(map[string]any)
	if !ok || labels["branch"] != "main" {
		t.Errorf("sent labels = %v, want branch=main", sent["labels"])
	}
	removed, ok := sent["removeLabels"].([]any)
	if !ok || len(removed) != 1 || removed[0] != "wip" {
		t.Errorf("sent removeLabels = %v, want [wip]", sent["removeLabels"])
	}
	if _, ok := sent["name"]; ok {
		t.Errorf("name should be omitted, body = %s", rs.lastBody)
	}
	if sess.Labels["branch"] != "main" {
		t.Errorf("response labels = %v, want branch=main", sess.Labels)
	}
}

func TestRestoreSession_Path(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"s1","status":"SESSION_STATUS_STARTING"}}`)

	if _, err := rs.client().RestoreSession(context.Background(), "ws-1", "s1"); err != nil {
		t.Fatalf("RestoreSession: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/restore"; rs.lastPath != want {
		t.Errorf("path = %s, want %s (canonical /restore, not /start)", rs.lastPath, want)
	}
}

func TestTerminateSession_Path(t *testing.T) {
	rs := newRecordingServer(t, `{"session":{"id":"s1","status":"SESSION_STATUS_TERMINATED"}}`)

	if _, err := rs.client().TerminateSession(context.Background(), "ws-1", "s1"); err != nil {
		t.Fatalf("TerminateSession: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/terminate"; rs.lastPath != want {
		t.Errorf("path = %s, want %s (canonical /terminate, not /stop)", rs.lastPath, want)
	}
}

func TestDeleteSession_Path(t *testing.T) {
	rs := newRecordingServer(t, ``)

	if err := rs.client().DeleteSession(context.Background(), "ws-1", "s1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if rs.lastMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
}

func TestDeleteTerminatedSessions_PathAndCount(t *testing.T) {
	rs := newRecordingServer(t, `{"deletedCount":3}`)

	n, err := rs.client().DeleteTerminatedSessions(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("DeleteTerminatedSessions: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions:delete-terminated"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if n != 3 {
		t.Errorf("deleted count = %d, want 3", n)
	}
}

func TestCompareSessionTemplate_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{
		"snapshot":{"templateName":"tmpl","stackId":"osx-xcode-16.0.x-stable"},
		"current":{"templateName":"tmpl","stackId":"osx-xcode-16.0.x-edge"},
		"changedVariableKeys":["FOO"]
	}`)

	resp, err := rs.client().CompareSessionTemplate(context.Background(), "ws-1", "s1")
	if err != nil {
		t.Fatalf("CompareSessionTemplate: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/template-diff"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if resp.Snapshot == nil || resp.Current == nil {
		t.Fatalf("snapshot/current missing: %+v", resp)
	}
	if resp.Snapshot.StackID != "osx-xcode-16.0.x-stable" || resp.Current.StackID != "osx-xcode-16.0.x-edge" {
		t.Errorf("stacks: snapshot=%q current=%q", resp.Snapshot.StackID, resp.Current.StackID)
	}
	if len(resp.ChangedVariableKeys) != 1 || resp.ChangedVariableKeys[0] != "FOO" {
		t.Errorf("changed keys = %+v", resp.ChangedVariableKeys)
	}
}

// TestSessions_ValidationGuards confirms every session method validates its
// required IDs before issuing an HTTP request.
func TestSessions_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	cases := map[string]func() error{
		"ListSessions/no-ws":           func() error { _, err := c.ListSessions(ctx, "", nil, ""); return err },
		"GetSession/no-ws":             func() error { _, err := c.GetSession(ctx, "", "s1"); return err },
		"GetSession/no-session":        func() error { _, err := c.GetSession(ctx, "ws", ""); return err },
		"CreateSession/no-ws":          func() error { _, _, err := c.CreateSession(ctx, "", CreateSessionRequest{}); return err },
		"UpdateSession/no-ws":          func() error { _, err := c.UpdateSession(ctx, "", "s1", UpdateSessionRequest{}); return err },
		"UpdateSession/no-session":     func() error { _, err := c.UpdateSession(ctx, "ws", "", UpdateSessionRequest{}); return err },
		"RestoreSession/no-session":    func() error { _, err := c.RestoreSession(ctx, "ws", ""); return err },
		"TerminateSession/no-session":  func() error { _, err := c.TerminateSession(ctx, "ws", ""); return err },
		"DeleteSession/no-session":     func() error { return c.DeleteSession(ctx, "ws", "") },
		"DeleteTerminated/no-ws":       func() error { _, err := c.DeleteTerminatedSessions(ctx, ""); return err },
		"CompareTemplate/no-session":   func() error { _, err := c.CompareSessionTemplate(ctx, "ws", ""); return err },
		"CompareTemplate/no-workspace": func() error { _, err := c.CompareSessionTemplate(ctx, "", "s1"); return err },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
	if rs.hits != 0 {
		t.Errorf("validation guards made %d HTTP call(s); should short-circuit", rs.hits)
	}
}
