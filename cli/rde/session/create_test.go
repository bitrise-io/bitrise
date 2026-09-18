package session

import (
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
)

func TestCreateCmd_RequiresTemplateAndName(t *testing.T) {
	// NAME given positionally but no --template fails fast.
	_, _, err := cmdtest.Run(t, newCreateCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      "http://unused",
		DefaultWorkspaceID: "ws-1",
		Args:               []string{"dev"},
	})
	if err == nil || !strings.Contains(err.Error(), "--template") {
		t.Errorf("error = %v, want --template required", err)
	}
}

func TestParseSessionInputs(t *testing.T) {
	got, err := parseSessionInputs(
		[]string{"repo=my-app"},
		[]string{"token=ghp_x"},
		[]string{"key=saved-id"},
	)
	if err != nil {
		t.Fatalf("parseSessionInputs: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d inputs, want 3", len(got))
	}
	if got[0].Key != "repo" || got[0].Value != "my-app" || got[0].IsSecret {
		t.Errorf("plain input wrong: %+v", got[0])
	}
	if !got[1].IsSecret || got[1].Value != "ghp_x" {
		t.Errorf("secret input wrong: %+v", got[1])
	}
	if got[2].SavedInputID != "saved-id" || got[2].Value != "" {
		t.Errorf("saved input wrong: %+v", got[2])
	}

	if _, err := parseSessionInputs([]string{"bad"}, nil, nil); err == nil {
		t.Error("expected error for malformed --input")
	}
	if _, err := parseSessionInputs(nil, nil, []string{"key="}); err == nil {
		t.Error("expected error for --saved-input with empty ID")
	}
}
