// Package style holds the lipgloss-backed semantic styles used by human
// renderers. JSON/YML output never goes through this package — ANSI codes
// must not leak into machine-readable output.
//
// This is a trimmed port of bitrise-cli's internal/output/style package:
// only the styles and Table renderer actually consumed by a landed command
// are included. Theme overrides, --no-color/--theme flags, and styles used
// only by commands not yet ported (OAuth picker, etc.) are left out on
// purpose — add them when a command that actually needs them lands, rather
// than carrying dead code now.
package style

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BrandColor is Bitrise's brand purple, used by the build-watch spinner.
const BrandColor = lipgloss.Color("#7B61FF")

// 256-color palette, paired by terminal background brightness. Each
// AdaptiveColor pairs a color tuned for dark terminals with one tuned for
// light terminals; lipgloss picks the appropriate side automatically.
var (
	dimColor     = lipgloss.AdaptiveColor{Light: "240", Dark: "245"} // grey
	successColor = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}   // green
	warnColor    = lipgloss.AdaptiveColor{Light: "136", Dark: "220"} // yellow / olive
	failedColor  = lipgloss.AdaptiveColor{Light: "160", Dark: "203"} // red
	runningColor = lipgloss.AdaptiveColor{Light: "25", Dark: "39"}   // blue
	abortedColor = lipgloss.AdaptiveColor{Light: "166", Dark: "208"} // orange
)

// Styles bundles the semantic styles used across human renderers. It is
// constructed per-writer so the writer's terminal capabilities (or lack
// thereof) control whether ANSI is emitted.
type Styles struct {
	Header  lipgloss.Style // table header rows
	Dim     lipgloss.Style // de-emphasized text
	Slug    lipgloss.Style // technical identifiers (dimmed)
	Success lipgloss.Style // success indicators
	Warn    lipgloss.Style // warnings
	Bold    lipgloss.Style // emphasis
	Label   lipgloss.Style // field labels in key/value dumps
	URL     lipgloss.Style // links

	failed  lipgloss.Style
	running lipgloss.Style
	aborted lipgloss.Style
}

// New returns a Styles bundle for the given writer. ANSI escape codes are
// emitted only if w is detected as a color-capable TTY (NO_COLOR is honored
// automatically by the underlying termenv detection).
func New(w io.Writer) Styles {
	r := lipgloss.NewRenderer(w)
	return Styles{
		Header:  r.NewStyle().Bold(true).Foreground(dimColor),
		Dim:     r.NewStyle().Foreground(dimColor),
		Slug:    r.NewStyle().Foreground(dimColor),
		Success: r.NewStyle().Foreground(successColor),
		Warn:    r.NewStyle().Foreground(warnColor),
		Bold:    r.NewStyle().Bold(true),
		Label:   r.NewStyle().Bold(true),
		URL:     r.NewStyle().Foreground(runningColor).Underline(true),
		failed:  r.NewStyle().Foreground(failedColor),
		running: r.NewStyle().Foreground(runningColor),
		aborted: r.NewStyle().Foreground(abortedColor),
	}
}

// BuildStatus returns the style to render a build status string in
// (success/failed/in-progress/aborted/aborted-with-success), falling back to
// Dim for any other value.
func (s Styles) BuildStatus(status string) lipgloss.Style {
	switch status {
	case "success":
		return s.Success
	case "failed":
		return s.failed
	case "in-progress":
		return s.running
	case "aborted", "aborted-with-success":
		return s.aborted
	default:
		return s.Dim
	}
}

// CellStyler renders a single cell content string. Only called for data rows
// (row >= 0); header cells are styled uniformly by Table using hdrStyle.
// Return the styled string; do not change the visible width.
type CellStyler func(row, col int, content string) string

// Table renders a borderless table with two-space gutters. Column widths are
// computed using lipgloss.Width so ANSI codes don't break alignment.
//
// hdrStyle is applied to header cells uniformly. cellStyler may be nil to
// emit unstyled cells.
func Table(w io.Writer, headers []string, rows [][]string, hdrStyle lipgloss.Style, cellStyler CellStyler) error {
	cols := len(headers)
	if cols == 0 {
		return nil
	}
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = lipgloss.Width(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if w := lipgloss.Width(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var sb strings.Builder
	for i, h := range headers {
		sb.WriteString(hdrStyle.Render(h))
		if i < cols-1 {
			sb.WriteString(strings.Repeat(" ", widths[i]-lipgloss.Width(h)+2))
		}
	}
	sb.WriteRune('\n')

	for r, row := range rows {
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			styled := cell
			if cellStyler != nil {
				styled = cellStyler(r, i, cell)
			}
			sb.WriteString(styled)
			if i < cols-1 {
				sb.WriteString(strings.Repeat(" ", widths[i]-lipgloss.Width(cell)+2))
			}
		}
		sb.WriteRune('\n')
	}

	_, err := io.WriteString(w, sb.String())
	return err
}
