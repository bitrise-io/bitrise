package rdeapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListSavedInputs_IsUserScoped(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInputs":[{"id":"sv-1","key":"gh","isSecret":true,"value":"***"}]}`)

	inputs, err := rs.client().ListSavedInputs(context.Background())
	if err != nil {
		t.Fatalf("ListSavedInputs: %v", err)
	}
	// User-scoped: no /workspaces/{id} segment.
	if want := "/v1/saved-inputs"; rs.lastPath != want {
		t.Errorf("path = %s, want %s (saved inputs are user-scoped)", rs.lastPath, want)
	}
	if len(inputs) != 1 || !inputs[0].IsSecret {
		t.Errorf("inputs = %+v", inputs)
	}
}

func TestGetSavedInput_Path(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInput":{"id":"sv-1","key":"gh"}}`)

	in, err := rs.client().GetSavedInput(context.Background(), "sv-1")
	if err != nil {
		t.Fatalf("GetSavedInput: %v", err)
	}
	if want := "/v1/saved-inputs/sv-1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if in.Key != "gh" {
		t.Errorf("key = %q, want gh", in.Key)
	}
}

func TestCreateSavedInput_BodyAndPath(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInput":{"id":"sv-new","key":"gh","isSecret":true}}`)

	in, err := rs.client().CreateSavedInput(context.Background(), CreateSavedInputRequest{
		Key:      "gh",
		Value:    "ghp_secret",
		IsSecret: true,
	})
	if err != nil {
		t.Fatalf("CreateSavedInput: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/saved-inputs"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent CreateSavedInputRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.Key != "gh" || sent.Value != "ghp_secret" || !sent.IsSecret {
		t.Errorf("sent = %+v", sent)
	}
	if in.ID != "sv-new" {
		t.Errorf("id = %q, want sv-new", in.ID)
	}
}

func TestUpdateSavedInput_OmitsUnsetPointerFields(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInput":{"id":"sv-1","key":"gh"}}`)

	value := "new-value"
	if _, err := rs.client().UpdateSavedInput(context.Background(), "sv-1", UpdateSavedInputRequest{Value: &value}); err != nil {
		t.Fatalf("UpdateSavedInput: %v", err)
	}
	if rs.lastMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", rs.lastMethod)
	}
	if want := "/v1/saved-inputs/sv-1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["value"] != "new-value" {
		t.Errorf("value = %v, want new-value", sent["value"])
	}
	// isSecret pointer is nil → must drop out so it isn't reset to false.
	if _, ok := sent["isSecret"]; ok {
		t.Errorf("isSecret should be omitted, body = %s", rs.lastBody)
	}
}

func TestUpdateSavedInput_FalseSecretIsSent(t *testing.T) {
	rs := newRecordingServer(t, `{"savedInput":{"id":"sv-1","key":"gh"}}`)

	secret := false
	if _, err := rs.client().UpdateSavedInput(context.Background(), "sv-1", UpdateSavedInputRequest{IsSecret: &secret}); err != nil {
		t.Fatalf("UpdateSavedInput: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// A non-nil pointer to false MUST be sent — that's the whole reason the
	// field is a *bool rather than a bool.
	v, ok := sent["isSecret"]
	if !ok {
		t.Fatalf("isSecret should be present (pointer to false), body = %s", rs.lastBody)
	}
	if v != false {
		t.Errorf("isSecret = %v, want false", v)
	}
}

func TestDeleteSavedInput_Path(t *testing.T) {
	rs := newRecordingServer(t, ``)

	if err := rs.client().DeleteSavedInput(context.Background(), "sv-1"); err != nil {
		t.Fatalf("DeleteSavedInput: %v", err)
	}
	if rs.lastMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", rs.lastMethod)
	}
	if want := "/v1/saved-inputs/sv-1"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
}

func TestSavedInputs_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	cases := map[string]func() error{
		"Get/no-id":    func() error { _, err := c.GetSavedInput(ctx, ""); return err },
		"Update/no-id": func() error { _, err := c.UpdateSavedInput(ctx, "", UpdateSavedInputRequest{}); return err },
		"Delete/no-id": func() error { return c.DeleteSavedInput(ctx, "") },
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
