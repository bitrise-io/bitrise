//go:build steplib_e2e

package steplibe2e

import (
	"encoding/json"
	"regexp"
	"strings"
)

// logLine is a single structured log message emitted by the CLI in JSON output
// mode. Message is normalized (used for matching/dedup); Raw is the original
// text with only ANSI stripped (shown in the report so diffs carry real values).
type logLine struct {
	Level   string
	Message string
	Raw     string
}

// rawJSONLine mirrors the CLI's JSON log schema (log/corelog/json_logger.go).
type rawJSONLine struct {
	Type     string `json:"type"`
	Producer string `json:"producer"`
	Level    string `json:"level"`
	Message  string `json:"message"`
}

// parseCLILogs extracts the bitrise_cli log messages from a JSON-formatted CLI
// run. Step output (producer "step") and non-log event lines are dropped: we
// compare the activation logs the CLI itself emits, which is where the stepman
// v1/v2 logger calls surface. Lines that are not valid JSON (e.g. a step's raw
// stdout that leaked through) are ignored.
func parseCLILogs(raw string) []logLine {
	var out []logLine
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var r rawJSONLine
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		if r.Type != "log" || r.Producer != "bitrise_cli" {
			continue
		}
		if isNoise(r.Message) {
			continue
		}
		msg := normalizeMessage(r.Message)
		if msg == "" {
			continue
		}
		out = append(out, logLine{Level: r.Level, Message: msg, Raw: rawText(r.Message)})
	}
	return out
}

// rawText strips ANSI and trailing whitespace but keeps the real values
// (paths, durations, URLs) so the report shows actual log content.
func rawText(s string) string {
	return strings.TrimSpace(reANSI.ReplaceAllString(s, ""))
}

// isNoise drops cosmetic and non-activation lines that would swamp the diff:
// the box-drawing run summary tables, blank lines, and the giant
// "Spec read from YML: models.StepModel{...}" struct dump.
func isNoise(msg string) bool {
	t := strings.TrimSpace(reANSI.ReplaceAllString(msg, ""))
	if t == "" {
		return true
	}
	if strings.HasPrefix(t, "Spec read from YML:") {
		return true
	}
	// Box-drawing / separator rows and the cosmetic step + run-summary table
	// rows ("| ✓ | … | 4.14 sec |", "| Total runtime: … |"). The stepman
	// activation logs we care about are plain messages, never table rows.
	if strings.HasPrefix(t, "|") || strings.HasPrefix(t, "+--") {
		return true
	}
	if strings.Trim(t, "+-| ") == "" {
		return true
	}
	return false
}

var (
	reANSI     = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	reDuration = regexp.MustCompile(`\b\d+(\.\d+)?(ns|µs|us|ms|s|m|h)\b`)
	reAbsPath  = regexp.MustCompile(`(/[^\s"']+)+`)
	reHexHash  = regexp.MustCompile(`\b[0-9a-f]{12,64}\b`)
	reWS       = regexp.MustCompile(`\s+`)
)

// normalizeMessage strips run-to-run variable substrings so the same logical
// message from two runs compares equal: ANSI colors, durations, absolute/temp
// paths, and long hex hashes. Step version numbers are intentionally NOT
// stripped — same-version cells share them, so they should match.
func normalizeMessage(s string) string {
	s = reANSI.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = reDuration.ReplaceAllString(s, "<dur>")
	s = reAbsPath.ReplaceAllString(s, "<path>")
	s = reHexHash.ReplaceAllString(s, "<hash>")
	s = reWS.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}

// key is the dedup/diff key for a log line: level + normalized message.
func (l logLine) key() string { return l.Level + "|" + l.Message }
