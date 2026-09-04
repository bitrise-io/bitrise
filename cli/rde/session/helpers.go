package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// validateLabelSelectors rejects selector shapes that can never match (bare
// keys, empty keys or values — label values are non-empty by contract). The
// strings otherwise pass to the backend verbatim, which enforces the full
// rules (at most 8 selectors, no duplicate keys).
func validateLabelSelectors(selectors []string) error {
	for _, sel := range selectors {
		k, v, ok := strings.Cut(sel, "=")
		if !ok || k == "" || v == "" {
			return fmt.Errorf("--label-selector %q: expected key=value", sel)
		}
	}
	return nil
}

// formatTime is the canonical timestamp format for session human output.
// Empty pointer renders as "" so renderers can shove it directly into a
// table cell without conditionals.
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

// statusStyle picks the right semantic color for a session status string.
// The status comes from internal/rde — already lowercased, prefix-stripped.
func statusStyle(s style.Styles, status string) lipgloss.Style {
	switch status {
	case "running":
		return s.BuildStatus("success")
	case "pending", "starting", "terminating", "draining":
		return s.BuildStatus("in-progress")
	case "terminated", "drained":
		return s.Dim
	case "failed":
		return s.BuildStatus("failed")
	}
	return s.Dim
}

// diskStatusText renders a terminated session's persistent-disk status with
// a hint about whether the session can still be restored. Callers guard on a
// non-empty status, so the default branch only fires for a value added to
// the backend after this code was written.
func diskStatusText(s style.Styles, status string) string {
	switch status {
	case internalrde.DiskStatusAvailable:
		return s.Success.Render("available") + s.Dim.Render(" — restorable")
	case internalrde.DiskStatusUnavailableSoon:
		return s.Warn.Render("expiring soon") + s.Dim.Render(" — restore within ~1 week")
	case internalrde.DiskStatusUnavailable:
		return s.Failure.Render("unavailable") + s.Dim.Render(" — cannot be restored")
	}
	return status
}
