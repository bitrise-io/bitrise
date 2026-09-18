package rdeapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListTemplates_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"templates":[
		{"id":"t1","name":"Linux Dev","stackId":"linux-ubuntu-24.04","machineType":"m1"},
		{"id":"t2","name":"macOS Dev","stackId":"osx-xcode-16.0.x-edge","machineType":"m2"}
	]}`)

	tmpls, err := rs.client().ListTemplates(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if want := "/v1/workspaces/ws-1/templates"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(tmpls) != 2 || tmpls[1].Name != "macOS Dev" {
		t.Errorf("templates = %+v", tmpls)
	}
}

func TestGetTemplate_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"template":{"id":"t1","name":"Dev","templateVariables":[{"key":"FOO","isSecret":true}]}}`)

	tmpl, err := rs.client().GetTemplate(context.Background(), "ws-1", "t1")
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if want := "/v1/workspaces/ws-1/templates/t1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(tmpl.TemplateVariables) != 1 || !tmpl.TemplateVariables[0].IsSecret {
		t.Errorf("template variables = %+v", tmpl.TemplateVariables)
	}
}

func TestCreateTemplate_BodyAndPath(t *testing.T) {
	rs := newRecordingServer(t, `{"template":{"id":"t-new","name":"Dev"}}`)

	tmpl, err := rs.client().CreateTemplate(context.Background(), "ws-1", CreateTemplateRequest{
		Name:        "Dev",
		StackID:     "osx-xcode-16.0.x-edge",
		MachineType: "g2.mac",
		SessionInputs: []SessionInputCreate{
			{Key: "repo", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/templates"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent CreateTemplateRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.StackID != "osx-xcode-16.0.x-edge" || sent.MachineType != "g2.mac" {
		t.Errorf("sent = %+v", sent)
	}
	if len(sent.SessionInputs) != 1 || !sent.SessionInputs[0].Required {
		t.Errorf("sent session inputs = %+v", sent.SessionInputs)
	}
	if tmpl.ID != "t-new" {
		t.Errorf("id = %q, want t-new", tmpl.ID)
	}
}

func TestUpdateTemplate_OmitsUnsetAndCarriesReplaceFlags(t *testing.T) {
	rs := newRecordingServer(t, `{"template":{"id":"t1","name":"Renamed"}}`)

	name := "Renamed"
	if _, err := rs.client().UpdateTemplate(context.Background(), "ws-1", "t1", UpdateTemplateRequest{
		Name:                &name,
		SessionInputs:       []SessionInputCreate{{Key: "repo"}},
		UpdateSessionInputs: true,
	}); err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}
	if rs.lastMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", rs.lastMethod)
	}
	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["name"] != "Renamed" {
		t.Errorf("name = %v, want Renamed", sent["name"])
	}
	// Unset scalar pointers drop out entirely.
	if _, ok := sent["stackId"]; ok {
		t.Errorf("stackId should be omitted, body = %s", rs.lastBody)
	}
	// The replace-flag must accompany the array it gates.
	if sent["updateSessionInputs"] != true {
		t.Errorf("updateSessionInputs = %v, want true", sent["updateSessionInputs"])
	}
	// Flags for arrays that weren't sent must stay omitted.
	if _, ok := sent["updateFeatureFlags"]; ok {
		t.Errorf("updateFeatureFlags should be omitted, body = %s", rs.lastBody)
	}
}

func TestDeleteTemplate_Path(t *testing.T) {
	rs := newRecordingServer(t, ``)

	if err := rs.client().DeleteTemplate(context.Background(), "ws-1", "t1"); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}
	if rs.lastMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/templates/t1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
}

func TestTemplates_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	cases := map[string]func() error{
		"List/no-ws":         func() error { _, err := c.ListTemplates(ctx, ""); return err },
		"Get/no-ws":          func() error { _, err := c.GetTemplate(ctx, "", "t1"); return err },
		"Get/no-template":    func() error { _, err := c.GetTemplate(ctx, "ws", ""); return err },
		"Create/no-ws":       func() error { _, err := c.CreateTemplate(ctx, "", CreateTemplateRequest{}); return err },
		"Update/no-ws":       func() error { _, err := c.UpdateTemplate(ctx, "", "t1", UpdateTemplateRequest{}); return err },
		"Update/no-template": func() error { _, err := c.UpdateTemplate(ctx, "ws", "", UpdateTemplateRequest{}); return err },
		"Delete/no-ws":       func() error { return c.DeleteTemplate(ctx, "", "t1") },
		"Delete/no-template": func() error { return c.DeleteTemplate(ctx, "ws", "") },
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
