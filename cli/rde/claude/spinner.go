package claude

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// await runs fn while showing an animated spinner line for a long, otherwise
// silent wait (e.g. the session boot). The spinner shows the elapsed time and
// an optional live sub-status that fn reports through the status callback. On
// success it leaves a settled "✓ <settled> (<elapsed>)" line in the scrollback;
// on failure it returns fn's error and prints nothing extra.
//
// It degrades to a plain static step line — today's behavior — whenever the
// output isn't an interactive color terminal or --quiet is set, so piped,
// redirected and CI output stays animation-free. Like the rest of stepLogger
// it only ever writes to stderr.
//
// await is synchronous: the spinner program is fully torn down (and the
// terminal restored) before it returns, so a following step that takes over
// the terminal — the git clone or the Claude Code attach — starts from a
// clean slate.
func (l *stepLogger) await(ctx context.Context, label, settled string, fn func(ctx context.Context, status func(string)) error) error {
	if l.quiet || !cmdutil.IsTerminalWriter(l.w) || !l.s.HasColor() {
		l.step("%s", label)
		return fn(ctx, func(string) {})
	}

	// A child context so cancelling here (e.g. on a render failure) also stops
	// fn, and an outer cancel (Ctrl-C) stops both the program and fn together.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()
	p := tea.NewProgram(
		newSpinnerModel(label, l.s),
		tea.WithContext(ctx),
		// No stdin: the spinner needs no keys, and reading stdin here could
		// swallow input meant for the step that follows. Ctrl-C still works
		// through the command's signal-aware context.
		tea.WithInput(bytes.NewReader(nil)),
		tea.WithOutput(l.w),
	)

	errCh := make(chan error, 1)
	go func() {
		err := fn(ctx, func(status string) { p.Send(statusMsg(status)) })
		p.Send(doneMsg{err: err})
		errCh <- err
	}()

	_, runErr := p.Run()
	cancel()
	err := <-errCh
	if err != nil {
		return err
	}
	// The wait succeeded; a render failure (rare) just means no settled line.
	if runErr == nil {
		l.settled(settled, time.Since(start))
	}
	return nil
}

// settled prints the indented, green-checked confirmation a finished await
// leaves behind: "  ✓ Session booted (1m31s)".
func (l *stepLogger) settled(label string, d time.Duration) {
	if l.quiet {
		return
	}
	_, _ = fmt.Fprintf(l.w, "%s%s %s\n", stepIndent,
		l.s.Success.Render("✓"),
		l.s.Dim.Render(fmt.Sprintf("%s (%s)", label, d.Round(time.Second))))
}

// statusMsg updates the spinner's live sub-status (e.g. "starting").
type statusMsg string

// doneMsg ends the spinner once the awaited work returns.
type doneMsg struct{ err error }

// tickMsg drives a once-a-second redraw so the elapsed counter advances even
// while the awaited work is quiet.
type tickMsg time.Time

func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// spinnerModel renders a single in-progress line: a brand-purple spinner, a
// label, the elapsed time, and an optional live sub-status. It collapses to ""
// once finished so the caller's settled line is what stays in the scrollback —
// the same disappearing-frame approach the picker uses.
type spinnerModel struct {
	spinner   spinner.Model
	label     string
	status    string
	startedAt time.Time
	finished  bool
	s         style.Styles
}

func newSpinnerModel(label string, s style.Styles) spinnerModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = s.Brand
	return spinnerModel{
		spinner:   sp,
		label:     label,
		startedAt: time.Now(),
		s:         s,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, elapsedTick())
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tickMsg:
		if m.finished {
			return m, nil
		}
		return m, elapsedTick()
	case statusMsg:
		m.status = string(msg)
		return m, nil
	case doneMsg:
		m.finished = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.finished {
		return ""
	}
	var b strings.Builder
	b.WriteString(stepIndent)
	b.WriteString(m.spinner.View()) // brand-purple frame, includes a trailing space
	b.WriteString(m.label)
	b.WriteString(m.s.Dim.Render("  " + m.elapsed()))
	if m.status != "" {
		b.WriteString(m.s.Dim.Render("  (" + m.status + ")"))
	}
	return b.String()
}

func (m spinnerModel) elapsed() string {
	return time.Since(m.startedAt).Round(time.Second).String()
}
