package rde

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
)

// DefaultMetadataInterval is how often ClaudeMetadataMonitor polls the session
// for the AI-generated title. One minute per the RDE plan: frequent enough that
// a freshly-named session shows a useful title soon, cheap enough to ignore.
const DefaultMetadataInterval = time.Minute

// ClaudeMetadataMonitor periodically reads the AI-generated title from the
// Claude Code transcript running inside an RDE session and, whenever the title
// or description changes, persists it to the local session store (so
// `rde claude --resume` has something descriptive to show) and pushes it to the
// API (so the session is recognizable in `rde session list` / the web UI).
//
// Everything is best-effort: a failed SSH read or API call is logged via Debug
// (if set) and retried on the next tick. It never disrupts the foreground
// Claude session.
type ClaudeMetadataMonitor struct {
	Service         *Service
	WorkspaceID     string
	SessionID       string
	ClaudeSessionID string
	Interval        time.Duration

	// Record is the monitor's own copy of the local record, taken at
	// construction: it updates and re-saves it as the title/description evolve.
	// The launcher never reads it back, so no synchronization is needed.
	Record localsession.Record

	// Describe returns the current session description (e.g.
	// "owner/repo @ branch" with the pull-request URL on its own line). Called
	// each tick because parts of it (the pull request) can appear after the
	// session starts. May be nil.
	Describe func(context.Context) string

	// Debug, if set, receives diagnostic messages about skipped/failed updates.
	Debug func(format string, args ...any)
}

// Run polls until ctx is cancelled. It checks once immediately (so the
// description is pushed promptly) and then on every Interval tick.
func (m *ClaudeMetadataMonitor) Run(ctx context.Context) {
	interval := m.Interval
	if interval <= 0 {
		interval = DefaultMetadataInterval
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		m.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (m *ClaudeMetadataMonitor) tick(ctx context.Context) {
	aiTitle := m.readAITitle(ctx)
	desc := ""
	if m.Describe != nil {
		desc = m.Describe(ctx)
	}

	updated, patch, changed := metadataUpdate(m.Record, aiTitle, desc)
	if !changed {
		return
	}
	m.Record = updated

	if err := localsession.Save(m.Record); err != nil {
		m.debugf("save local session record: %v", err)
	}
	if _, err := m.Service.UpdateSession(ctx, m.WorkspaceID, m.SessionID, patch); err != nil {
		m.debugf("update session metadata: %v", err)
	}
}

// readAITitle reads the latest AI-generated title from the Claude transcript on
// the session. Returns "" when the transcript or a title isn't there yet.
func (m *ClaudeMetadataMonitor) readAITitle(ctx context.Context) string {
	if m.ClaudeSessionID == "" {
		return ""
	}
	res, err := m.Service.Execute(ctx, m.WorkspaceID, m.SessionID, readAITitleCommand(m.ClaudeSessionID), nil, DefaultExecuteTimeout)
	if err != nil {
		m.debugf("read ai-title: %v", err)
		return ""
	}
	return parseAITitle(res.Stdout)
}

func (m *ClaudeMetadataMonitor) debugf(format string, args ...any) {
	if m.Debug != nil {
		m.Debug(format, args...)
	}
}

// readAITitleCommand returns a remote shell command that prints the last
// "ai-title" JSON line from the Claude transcript for claudeSessionID. The
// transcript is named <session-id>.jsonl and lives under one of the
// ~/.claude/projects/<project>/ directories; the session ID is unique, so the
// glob matches at most one file. grep|tail keeps it cheap — no full transfer.
//
// The ID is quoted even though localsession only ever loads UUIDs (see
// readRecord): quoting the filename segment still leaves the leading * to
// expand, so it costs nothing.
func readAITitleCommand(claudeSessionID string) string {
	return "f=$(ls -1 ~/.claude/projects/*/" + shellSingleQuote(claudeSessionID+".jsonl") + " 2>/dev/null | head -n1); " +
		"test -n \"$f\" && grep '\"type\":\"ai-title\"' \"$f\" | tail -n1"
}

// parseAITitle extracts aiTitle from the last "ai-title" JSONL record in out.
// Returns "" when out is empty or not such a record.
func parseAITitle(out string) string {
	line := lastNonBlankLine(out)
	if line == "" {
		return ""
	}
	var rec struct {
		Type    string `json:"type"`
		AITitle string `json:"aiTitle"`
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		return ""
	}
	if rec.Type != "ai-title" {
		return ""
	}
	return rec.AITitle
}

// metadataUpdate computes the record + API patch for a freshly-observed title
// and description. A title only overrides the name once it's non-empty (so the
// generated "claude-<hex>" name isn't clobbered before Claude names the
// session). Returns changed=false when nothing differs from the current record.
func metadataUpdate(rec localsession.Record, aiTitle, desc string) (localsession.Record, UpdateSessionRequest, bool) {
	var nameChanged, descChanged bool
	if aiTitle != "" && aiTitle != rec.AITitle {
		rec.AITitle = aiTitle
		nameChanged = true
	}
	if desc != "" && desc != rec.Description {
		rec.Description = desc
		descChanged = true
	}
	if !nameChanged && !descChanged {
		return rec, UpdateSessionRequest{}, false
	}
	var patch UpdateSessionRequest
	if nameChanged {
		n := aiTitle
		patch.Name = &n
	}
	if descChanged {
		d := desc
		patch.Description = &d
	}
	return rec, patch, true
}

func lastNonBlankLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}
