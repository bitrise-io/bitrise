//go:build steplib_e2e

package steplibe2e

import (
	"encoding/json"
	"strings"
)

// failureCase is a negative case: activation is expected to fail. The same case
// is run through the v1 and v2 paths so their error behavior can be compared.
type failureCase struct {
	name    string
	stepID  string
	version string // bad/missing version, or unparseable constraint
	desc    string
}

// failureCases covers the resolution-failure modes. git-clone is a real step, so
// these isolate version-resolution errors from "step missing" errors.
func failureCases() []failureCase {
	return []failureCase{
		{name: "step-id-not-found", stepID: "this-step-does-not-exist-xyz", version: "1.0.0", desc: "unknown step id"},
		{name: "invalid-version-constraint", stepID: "git-clone", version: "not-a-version", desc: "unparseable version constraint"},
		{name: "exact-version-not-found", stepID: "git-clone", version: "99.99.99", desc: "pinned exact version absent"},
		{name: "major-lock-not-found", stepID: "git-clone", version: "99", desc: "major-locked version absent"},
		{name: "minor-lock-not-found", stepID: "git-clone", version: "8.99", desc: "minor-locked version absent"},
	}
}

// cell adapts a failure case to a runnable cell (no inputs — activation is
// expected to fail before the step would run).
func (f failureCase) cell(variant pathVariant) cell {
	return cell{
		step:    stepSpec{id: f.stepID},
		version: versionForm{label: f.name, version: f.version},
		variant: variant,
	}
}

// failureRow is the per-case comparison of how v1 and v2 reported the failure.
type failureRow struct {
	name       string
	desc       string
	ref        string
	v1Failed   bool
	v1Message  string
	v2Failed   bool
	v2Message  string
}

// stepFinishedEvent mirrors the `step_finished` structured event the CLI emits
// in JSON mode. Its content carries the activation outcome and, on failure, the
// exact error message(s) — a far more robust signal than scanning colored or
// leveled log text (the cosmetic failure summary is only normal-level).
type stepFinishedEvent struct {
	Type      string `json:"type"`
	EventType string `json:"event_type"`
	Content   struct {
		Status string `json:"status"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"content"`
}

// stepFailure scans the captured output for a step_finished event whose status
// is not success, returning whether the step failed and the joined error
// message(s) from the event.
func stepFailure(raw string) (failed bool, message string) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var e stepFinishedEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if e.Type != "event" || e.EventType != "step_finished" {
			continue
		}
		if e.Content.Status == "" || e.Content.Status == "success" {
			continue
		}
		var msgs []string
		for _, er := range e.Content.Errors {
			msgs = append(msgs, er.Message)
		}
		msg := strings.Join(msgs, "; ")
		if len(msg) > 280 {
			msg = msg[:280] + "…"
		}
		return true, msg
	}
	return false, ""
}

// failureMessage describes why a failed cell failed: the structured
// step_finished error when present, else a panic line (the activation crashed
// rather than failing gracefully), else a generic note.
func failureMessage(r cellResult) string {
	if _, msg := stepFailure(r.output); msg != "" {
		return msg
	}
	for _, line := range strings.Split(r.output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "panic:") {
			if len(line) > 200 {
				line = line[:200] + "…"
			}
			return "crashed — " + line
		}
	}
	return "(failed without a structured error event)"
}
