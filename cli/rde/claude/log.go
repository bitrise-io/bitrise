package claude

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// stepLogger writes the command's grouped progress to stderr: a bold group
// title, indented step lines, and a green "✓" confirmation when a group
// completes. Warnings use the yellow Warn style. Everything is a diagnostic —
// it never touches stdout.
type stepLogger struct {
	w       io.Writer
	s       style.Styles
	quiet   bool
	grouped bool // whether a group has been printed yet (for separator spacing)
}

func newStepLogger(cmd *cobra.Command) *stepLogger {
	w := cmd.ErrOrStderr()
	return &stepLogger{w: w, s: style.New(w), quiet: cmdutil.IsQuiet(cmd)}
}

// group starts a new section with a bold title, separated from the previous
// group by a blank line (omitted before the first group).
func (l *stepLogger) group(title string) {
	if l.quiet {
		return
	}
	sep := "\n"
	if !l.grouped {
		sep = ""
		l.grouped = true
	}
	_, _ = fmt.Fprintf(l.w, "%s%s\n", sep, l.s.Bold.Render(title))
}

// step reports an in-progress action within the current group (indented, dim).
func (l *stepLogger) step(format string, a ...any) {
	if l.quiet {
		return
	}
	_, _ = fmt.Fprintf(l.w, "  %s\n", l.s.Dim.Render(fmt.Sprintf(format, a...)))
}

// done confirms a completed group with the green "✓" success marker.
func (l *stepLogger) done(format string, a ...any) {
	if l.quiet {
		return
	}
	_, _ = fmt.Fprintf(l.w, "%s %s\n", l.s.Success.Render("✓"), fmt.Sprintf(format, a...))
}

// warn emits a yellow warning. Unlike group/step/done, warnings are not
// suppressed by --quiet.
func (l *stepLogger) warn(format string, a ...any) {
	_, _ = fmt.Fprintf(l.w, "  %s\n", l.s.Warn.Render(fmt.Sprintf(format, a...)))
}

// stepIndent is the prefix step lines use; indentWriter reuses it so streamed
// remote output (e.g. git clone progress) lines up under its group.
const stepIndent = "  "

// indentWriter prefixes every line written through it with a fixed indent, so
// a remote command's streamed output aligns with the logger's step lines. It
// tracks line starts across writes and treats both "\n" and "\r" as line
// boundaries, so carriage-return progress redraws (git's "Receiving objects…")
// stay indented too. It is NOT suitable for full-screen TUIs that position the
// cursor absolutely (e.g. Claude Code itself) — only for line-oriented output.
type indentWriter struct {
	w           io.Writer
	prefix      string
	atLineStart bool
}

// newIndentWriter wraps w so each output line is prefixed with stepIndent.
func newIndentWriter(w io.Writer) *indentWriter {
	return &indentWriter{w: w, prefix: stepIndent, atLineStart: true}
}

func (iw *indentWriter) Write(p []byte) (int, error) {
	var buf []byte
	for _, b := range p {
		if iw.atLineStart && b != '\n' && b != '\r' {
			buf = append(buf, iw.prefix...)
			iw.atLineStart = false
		}
		buf = append(buf, b)
		if b == '\n' || b == '\r' {
			iw.atLineStart = true
		}
	}
	if _, err := iw.w.Write(buf); err != nil {
		return 0, err
	}
	return len(p), nil
}
