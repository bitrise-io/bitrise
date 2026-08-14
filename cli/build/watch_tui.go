package build

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// runWatchTUI is the interactive variant of runWatch. It renders a permanent
// status bar (spinner + build info + clickable URL) at the bottom of the
// terminal while build logs scroll above it. When the build finishes the TUI
// exits cleanly and the caller renders the final summary.
func runWatchTUI(cmd *cobra.Command, svc *internalbuild.Service, b internalbuild.Build, interval time.Duration) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	m := newWaitModel(b)
	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithOutput(cmd.OutOrStdout()))

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		finalBuild, err := svc.Watch(ctx, b.AppSlug, b.Slug, &teaLogWriter{p: p}, interval)
		p.Send(watchDoneMsg{build: finalBuild, err: err})
	}()

	finalModel, err := p.Run()
	cancel()
	<-doneCh
	if err != nil {
		return fmt.Errorf("render wait UI: %w", err)
	}

	fm, ok := finalModel.(waitModel)
	if !ok {
		return fmt.Errorf("unexpected final model type %T", finalModel)
	}

	stderr := cmd.ErrOrStderr()
	// User interrupted (Ctrl-C, SIGTERM) before the watch returned a result.
	// fm.finished is only true after watchDoneMsg arrived and was handled;
	// fm.finalErr is context.Canceled when Watch itself observed the cancel.
	if !fm.finished || errors.Is(fm.finalErr, context.Canceled) {
		return writeDetachNotice(stderr, "build watch "+b.Slug)
	}
	if fm.finalErr != nil {
		return fm.finalErr
	}

	final := fm.finalBuild
	footer := fmt.Sprintf("Build #%d finished: %s%s\n", final.BuildNumber, final.Status, buildElapsed(final))
	if url := buildDetailURL(cmd, b); url != "" {
		footer += fmt.Sprintf("%s\n", url)
	}
	if _, err := fmt.Fprint(stderr, footer); err != nil {
		return err
	}
	if final.Status != "success" && final.Status != "aborted-with-success" {
		return fmt.Errorf("build %s", final.Status)
	}
	return nil
}

// teaLogWriter adapts svc.Watch's io.Writer log sink to the bubbletea
// program. Each Write becomes a logChunkMsg the model splits into lines
// printed above the status bar via tea.Println.
type teaLogWriter struct {
	p *tea.Program
}

func (w *teaLogWriter) Write(p []byte) (int, error) {
	w.p.Send(logChunkMsg(string(p)))
	return len(p), nil
}

type logChunkMsg string

// printDoneMsg signals that the in-flight tea.Println sequence has finished,
// so the next batch of buffered log lines may be printed. See flushPending.
type printDoneMsg struct{}

type watchDoneMsg struct {
	build internalbuild.Build
	err   error
}

type tickMsg time.Time

type waitModel struct {
	build      internalbuild.Build
	spinner    spinner.Model
	leftover   string
	pending    []string
	printing   bool
	quitting   bool
	startedAt  time.Time
	finalBuild internalbuild.Build
	finalErr   error
	finished   bool
	width      int
	labelStyle lipgloss.Style
	dimStyle   lipgloss.Style
	urlStyle   lipgloss.Style
}

func newWaitModel(b internalbuild.Build) waitModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(style.BrandColor)
	started := b.TriggeredAt
	if started.IsZero() {
		started = time.Now()
	}
	return waitModel{
		build:      b,
		spinner:    sp,
		startedAt:  started,
		width:      80,
		labelStyle: lipgloss.NewStyle().Bold(true),
		dimStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		urlStyle:   lipgloss.NewStyle().Foreground(style.BrandColor).Underline(true),
	}
}

func (m waitModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, elapsedTick())
}

func (m waitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case logChunkMsg:
		m.leftover += string(msg)
		for {
			i := strings.IndexByte(m.leftover, '\n')
			if i < 0 {
				break
			}
			line := strings.TrimRight(m.leftover[:i], "\r")
			m.leftover = m.leftover[i+1:]
			m.pending = append(m.pending, line)
		}
		return m.flushPending()

	case printDoneMsg:
		m.printing = false
		if m.quitting {
			return m.flushFinal()
		}
		return m.flushPending()

	case watchDoneMsg:
		m.finalBuild = msg.build
		m.finalErr = msg.err
		m.finished = true
		m.quitting = true
		// Any trailing partial line (no terminating newline) is the last
		// thing to print.
		if m.leftover != "" {
			m.pending = append(m.pending, m.leftover)
			m.leftover = ""
		}
		// If a print is still in flight, wait for it; printDoneMsg will run
		// flushFinal once it completes. Otherwise flush + quit now.
		if m.printing {
			return m, nil
		}
		return m.flushFinal()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		if m.finished {
			return m, nil
		}
		return m, elapsedTick()
	}
	return m, nil
}

// flushPending prints all buffered complete lines as one ordered block, but
// only when no print is already in flight. Bubbletea runs commands returned
// from Update concurrently (tea.Batch makes no ordering promise, and even
// separate Update calls race), so emitting one tea.Println per line lets the
// scheduler interleave them — which is exactly the log scrambling we're
// fixing. Instead we serialize: at most one print command is outstanding,
// new lines accumulate in m.pending while it runs, and printDoneMsg releases
// the next block once the previous one lands.
func (m waitModel) flushPending() (tea.Model, tea.Cmd) {
	if m.printing || len(m.pending) == 0 {
		return m, nil
	}
	block := strings.Join(m.pending, "\n")
	m.pending = nil
	m.printing = true
	return m, tea.Sequence(
		tea.Println(block),
		func() tea.Msg { return printDoneMsg{} },
	)
}

// flushFinal prints any remaining buffered lines and quits. Called once the
// build is done and no print is in flight, so ordering is preserved.
func (m waitModel) flushFinal() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if len(m.pending) > 0 {
		cmds = append(cmds, tea.Println(strings.Join(m.pending, "\n")))
		m.pending = nil
	}
	cmds = append(cmds, tea.Quit)
	return m, tea.Sequence(cmds...)
}

func (m waitModel) View() string {
	if m.finished {
		return ""
	}

	info := strings.Builder{}
	info.WriteString(m.spinner.View())
	info.WriteString(" ")
	info.WriteString(m.labelStyle.Render("Building"))
	if m.build.BuildNumber > 0 {
		info.WriteString(m.dimStyle.Render(fmt.Sprintf(" #%d", m.build.BuildNumber)))
	}
	if m.build.Workflow != "" {
		info.WriteString("  ")
		info.WriteString(m.build.Workflow)
	}
	switch {
	case m.build.Branch != "":
		info.WriteString(m.dimStyle.Render(" on "))
		info.WriteString(m.build.Branch)
	case m.build.Tag != "":
		info.WriteString(m.dimStyle.Render(" tag "))
		info.WriteString(m.build.Tag)
	}
	info.WriteString(m.dimStyle.Render(fmt.Sprintf("  %s elapsed", m.elapsed())))

	if m.build.BuildURL == "" {
		return info.String()
	}
	// URL on its own line — most modern terminals make plain URLs Cmd-/Ctrl-
	// clickable, and a full-line URL avoids truncation on narrow terminals.
	url := m.dimStyle.Render("-> ") + m.urlStyle.Render(m.build.BuildURL)
	return info.String() + "\n" + url
}

func (m waitModel) elapsed() string {
	return time.Since(m.startedAt).Round(time.Second).String()
}

// elapsedTick redraws the status bar once per second so the elapsed-time
// counter advances even between log chunks.
func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
