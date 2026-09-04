package rde

import (
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
)

func TestParseAITitle(t *testing.T) {
	transcript := strings.Join([]string{
		`{"type":"mode","mode":"normal","sessionId":"s"}`,
		`{"type":"user","message":{"content":"hi"}}`,
		`{"type":"ai-title","aiTitle":"First title","sessionId":"s"}`,
		`{"type":"assistant","message":{}}`,
		`{"type":"ai-title","aiTitle":"Refined title","sessionId":"s"}`,
		"",
	}, "\n")
	if got := parseAITitle(transcript); got != "Refined title" {
		t.Errorf("parseAITitle = %q, want last ai-title", got)
	}

	if got := parseAITitle(""); got != "" {
		t.Errorf("empty input = %q, want empty", got)
	}
	// Last non-blank line is not an ai-title record.
	if got := parseAITitle(`{"type":"assistant"}`); got != "" {
		t.Errorf("non-ai-title last line = %q, want empty", got)
	}
	// Only the last line is inspected (grep|tail gives us one line in prod,
	// but be robust).
	if got := parseAITitle("garbage\n" + `{"type":"ai-title","aiTitle":"T"}`); got != "T" {
		t.Errorf("trailing ai-title = %q, want T", got)
	}
}

func TestReadAITitleCommand(t *testing.T) {
	cmd := readAITitleCommand("abc-123")
	if !strings.Contains(cmd, "abc-123.jsonl") {
		t.Errorf("command missing session-id glob: %q", cmd)
	}
	if !strings.Contains(cmd, `grep '"type":"ai-title"'`) {
		t.Errorf("command missing ai-title grep: %q", cmd)
	}
	if !strings.Contains(cmd, "~/.claude/projects/*/") {
		t.Errorf("command missing projects glob: %q", cmd)
	}
}

// localsession rejects a non-UUID session ID at load, so this is belt-and-
// braces: the ID still must not be able to break out of the glob and run a
// second command.
func TestReadAITitleCommandQuotesSessionID(t *testing.T) {
	cmd := readAITitleCommand("x.jsonl; touch /tmp/pwned; #")
	if strings.Contains(cmd, "; touch /tmp/pwned") && !strings.Contains(cmd, `'x.jsonl; touch /tmp/pwned; #.jsonl'`) {
		t.Errorf("session ID not quoted, command is injectable: %q", cmd)
	}
	// The leading * must still be left outside the quotes to expand.
	if !strings.Contains(cmd, "~/.claude/projects/*/'") {
		t.Errorf("quoting swallowed the project glob: %q", cmd)
	}
}

func TestMetadataUpdate(t *testing.T) {
	base := localsession.Record{Name: "claude-x"}

	// First observation: title + description both new.
	updated, patch, changed := metadataUpdate(base, "My Title", "org/repo @ main")
	if !changed {
		t.Fatal("expected changed")
	}
	if updated.AITitle != "My Title" || updated.Description != "org/repo @ main" {
		t.Errorf("record not updated: %+v", updated)
	}
	if patch.Name == nil || *patch.Name != "My Title" {
		t.Errorf("patch.Name = %v, want My Title", patch.Name)
	}
	if patch.Description == nil || *patch.Description != "org/repo @ main" {
		t.Errorf("patch.Description = %v", patch.Description)
	}

	// No change → not changed, empty patch.
	if _, _, changed := metadataUpdate(updated, "My Title", "org/repo @ main"); changed {
		t.Error("expected no change on identical input")
	}

	// Only description changes (PR appeared).
	_, patch, changed = metadataUpdate(updated, "My Title", "org/repo @ main (#7)")
	if !changed {
		t.Fatal("expected change when description changes")
	}
	if patch.Name != nil {
		t.Errorf("patch.Name should be nil when only description changed, got %v", *patch.Name)
	}
	if patch.Description == nil || *patch.Description != "org/repo @ main (#7)" {
		t.Errorf("patch.Description = %v", patch.Description)
	}

	// Empty title must not clobber the generated name.
	_, patch, changed = metadataUpdate(base, "", "org/repo @ main")
	if !changed {
		t.Fatal("expected change for new description")
	}
	if patch.Name != nil {
		t.Errorf("empty title set a name patch: %v", *patch.Name)
	}
}
