// Package picker is a small reusable arrow-key list selector built on
// bubbletea. It renders an inline (non-alt-screen) menu — a highlighted cursor
// you move with ↑/↓ (or j/k), Enter to confirm, Esc/q/Ctrl-C to cancel — and
// returns the chosen index. The frame collapses to "" on exit so the caller's
// own confirmation/log line is what stays in the scrollback.
//
// The picker is domain-agnostic: callers describe each row with an Item and map
// any per-row status color onto a Tone, so the picker never has to know about a
// caller's status vocabulary. Callers own the non-TTY fallback and the
// single-item shortcut; Select always goes interactive.
package picker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// Tone is a semantic color hint for an item's status text. It decouples the
// picker from any domain's status words: the caller maps its own statuses to a
// Tone, and the picker maps Tone → the shared style bundle.
type Tone int

const (
	ToneNone    Tone = iota // no status column
	ToneSuccess             // green
	ToneWarn                // amber
	ToneDanger              // red
	ToneDim                 // grey / de-emphasized
)

// Item is one row. Most items are selectable; set Divider to make a
// non-selectable group header instead.
type Item struct {
	Title  string // primary text (e.g. an image name or a session title)
	Desc   string // optional dim secondary text shown after the title
	Status string // optional right-side status word ("" = none)
	Tone   Tone   // color hint for Status
	// Divider marks a non-selectable separator row — a group header. The cursor
	// skips over it, Enter can't choose it, and it renders as a dim bold header
	// rather than a normal row. Only Title is used; the other fields are ignored.
	Divider bool
}

// Config controls a single Select call.
type Config struct {
	// Prompt heads the list, e.g. "Select an image".
	Prompt string
	// Note is an optional dim block rendered between the prompt and the rows,
	// e.g. a key/value summary of what's being chosen. May span multiple lines;
	// each line is rendered as-is (indent it yourself).
	Note  string
	Items []Item
	// Cursor is the index highlighted on open (the press-Enter choice).
	// Out-of-range values fall back to 0.
	Cursor int
	// DefaultIdx marks one row with a "(default)" badge (-1 = none).
	DefaultIdx int
	// Filter enables type-to-narrow substring filtering. When off, letter keys
	// drive j/k navigation and q quits.
	Filter bool
	// In is the key source (os.Stdin); Out is where the menu renders (stderr).
	In  io.Reader
	Out io.Writer
}

// ErrCancelled is returned when the user backs out (Esc, q, Ctrl-C/SIGTERM,
// EOF) or when there are no items to choose from. Callers translate it to their
// own clean-exit sentinel.
var ErrCancelled = errors.New("picker cancelled")

// Select runs the interactive picker and returns the chosen index into
// cfg.Items. It returns ErrCancelled if the user backs out (or cfg.Items is
// empty), and a wrapped error if the bubbletea program itself fails. The caller
// is responsible for the non-TTY fallback and the single-item shortcut — Select
// always goes interactive. ctx cancellation yields ErrCancelled.
func Select(ctx context.Context, cfg Config) (int, error) {
	if len(cfg.Items) == 0 {
		return 0, ErrCancelled
	}
	p := tea.NewProgram(newModel(cfg),
		tea.WithContext(ctx),
		tea.WithInput(cfg.In),   // os.Stdin
		tea.WithOutput(cfg.Out), // stderr — keeps stdout clean (data discipline)
	)
	final, err := p.Run()
	if err != nil {
		// External cancel (trapped Ctrl-C / SIGTERM via WithContext) surfaces
		// here; treat it as a clean backout like an in-program cancel.
		if errors.Is(err, context.Canceled) || errors.Is(err, tea.ErrProgramKilled) {
			return 0, ErrCancelled
		}
		return 0, fmt.Errorf("render picker: %w", err)
	}
	m, ok := final.(model)
	if !ok {
		return 0, fmt.Errorf("unexpected final model type %T", final)
	}
	if m.cancelled || m.chosen < 0 {
		return 0, ErrCancelled
	}
	return m.chosen, nil
}

// model is the bubbletea state. cursor/offset index into the *filtered* view;
// view holds the item indices currently shown.
type model struct {
	prompt   string
	note     string
	items    []Item
	filterOn bool

	view   []int  // item indices currently visible (filter applied)
	cursor int    // index into view
	offset int    // first visible view index (scroll window)
	filter string // current filter text

	defaultIdx int  // item index to badge "(default)" (-1 = none)
	chosen     int  // chosen item index; -1 until Enter
	cancelled  bool // user backed out
	width      int
	height     int

	s           style.Styles
	cursorStyle lipgloss.Style
}

func newModel(cfg Config) model {
	view := make([]int, len(cfg.Items))
	for i := range cfg.Items {
		view[i] = i
	}
	cursor := cfg.Cursor
	if cursor < 0 || cursor >= len(cfg.Items) {
		cursor = 0
	}
	def := cfg.DefaultIdx
	if def < 0 || def >= len(cfg.Items) {
		def = -1
	}
	s := style.New(cfg.Out)
	m := model{
		prompt:      cfg.Prompt,
		note:        cfg.Note,
		items:       cfg.Items,
		filterOn:    cfg.Filter,
		view:        view,
		cursor:      cursor,
		defaultIdx:  def,
		chosen:      -1,
		width:       80,
		s:           s,
		cursorStyle: s.Brand.Bold(true),
	}
	m.cursor = m.selectableCursor(cursor) // never open the cursor on a divider
	m.clampOffset()
	return m
}

// selectableCursor returns a view index that lands on a selectable (non-divider)
// row: it searches from want forward, then backward, so a group header is never
// the highlighted row. Falls back to want when every visible row is a divider
// (callers always include at least one selectable item, so this is defensive).
func (m model) selectableCursor(want int) int {
	if want < 0 || want >= len(m.view) {
		want = 0
	}
	for i := want; i < len(m.view); i++ {
		if !m.items[m.view[i]].Divider {
			return i
		}
	}
	for i := want - 1; i >= 0; i-- {
		if !m.items[m.view[i]].Divider {
			return i
		}
	}
	return want
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampOffset()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyUp:
			m.move(-1)
		case tea.KeyDown:
			m.move(1)
		case tea.KeyEnter:
			if len(m.view) > 0 && !m.items[m.view[m.cursor]].Divider {
				m.chosen = m.view[m.cursor]
				return m, tea.Quit
			}
		case tea.KeyBackspace:
			if m.filterOn && m.filter != "" {
				r := []rune(m.filter)
				m.filter = string(r[:len(r)-1])
				m.applyFilter()
			}
		case tea.KeySpace:
			if m.filterOn {
				m.filter += " "
				m.applyFilter()
			}
		case tea.KeyRunes:
			if m.filterOn {
				m.filter += string(msg.Runes)
				m.applyFilter()
				break
			}
			switch string(msg.Runes) {
			case "k":
				m.move(-1)
			case "j":
				m.move(1)
			case "q":
				m.cancelled = true
				return m, tea.Quit
			}
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.chosen >= 0 || m.cancelled {
		return "" // collapse — the caller prints the confirmation line
	}

	var b strings.Builder
	b.WriteString(m.s.Bold.Render(m.prompt))
	if m.filterOn && m.filter != "" {
		b.WriteString("  ")
		b.WriteString(m.s.Dim.Render("/" + m.filter))
	}
	b.WriteByte('\n')

	if m.note != "" {
		for _, line := range strings.Split(m.note, "\n") {
			b.WriteString(m.s.Dim.Render(line))
			b.WriteByte('\n')
		}
		b.WriteByte('\n') // blank line between the summary and the rows
	}

	if len(m.view) == 0 {
		b.WriteString(m.s.Dim.Render("  no matches"))
		b.WriteByte('\n')
		b.WriteString(m.s.Dim.Render(m.hint()))
		return b.String()
	}

	maxRows := m.visibleRows()
	start := m.offset
	end := start + maxRows
	if end > len(m.view) {
		end = len(m.view)
	}
	for vi := start; vi < end; vi++ {
		itemIdx := m.view[vi]
		it := m.items[itemIdx]
		if it.Divider {
			// A group header: flush-left, bold/dim, never carries the cursor.
			b.WriteString(m.s.Header.Render(it.Title))
			b.WriteByte('\n')
			continue
		}
		marker := "  "
		title := it.Title
		if vi == m.cursor {
			marker = m.cursorStyle.Render("›") + " "
			title = m.cursorStyle.Render(title)
		}
		line := marker + title
		if it.Desc != "" {
			line += "  " + m.s.Dim.Render(it.Desc)
		}
		if itemIdx == m.defaultIdx {
			line += "  " + m.s.Dim.Render("(default)")
		}
		if it.Status != "" {
			line += "  " + m.toneStyle(it.Tone).Render("("+it.Status+")")
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if len(m.view) > maxRows {
		b.WriteString(m.s.Dim.Render(fmt.Sprintf("  … %d–%d of %d", start+1, end, len(m.view))))
		b.WriteByte('\n')
	}
	b.WriteString(m.s.Dim.Render(m.hint()))
	return b.String()
}

func (m model) hint() string {
	if m.filterOn {
		return "↑/↓ move · type to filter · enter select · esc cancel"
	}
	return "↑/↓ move · enter select · esc cancel"
}

func (m model) toneStyle(t Tone) lipgloss.Style {
	switch t {
	case ToneSuccess:
		return m.s.Success
	case ToneWarn:
		return m.s.Warn
	case ToneDanger:
		return m.s.Failure
	default:
		return m.s.Dim
	}
}

// move shifts the cursor within the filtered view (clamped, no wrap), skipping
// over any divider rows, and keeps it inside the scroll window. When there's no
// selectable row further in the given direction the cursor stays put.
func (m *model) move(delta int) {
	if len(m.view) == 0 || delta == 0 {
		return
	}
	next := m.cursor
	for {
		next += delta
		if next < 0 || next >= len(m.view) {
			return // edge reached with only dividers beyond: stay put
		}
		if !m.items[m.view[next]].Divider {
			m.cursor = next
			m.clampOffset()
			return
		}
	}
}

// applyFilter rebuilds the visible view from a case-insensitive substring match
// over Title+Desc, then resets the cursor/window to the top match.
func (m *model) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(m.filter))
	view := make([]int, 0, len(m.items))
	for i, it := range m.items {
		if q == "" || strings.Contains(strings.ToLower(it.Title+" "+it.Desc), q) {
			view = append(view, i)
		}
	}
	m.view = view
	m.cursor = m.selectableCursor(0)
	m.offset = 0
}

// clampOffset scrolls the window so the cursor stays visible.
func (m *model) clampOffset() {
	if len(m.view) == 0 {
		m.offset = 0
		return
	}
	maxRows := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+maxRows {
		m.offset = m.cursor - maxRows + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > len(m.view)-1 {
		m.offset = len(m.view) - 1
	}
}

// visibleRows is how many item rows fit, reserving lines for the prompt, the
// "… N of M" window summary, and the footer hint. Falls back to a sane cap
// before the first WindowSizeMsg (height unknown).
func (m model) visibleRows() int {
	const reserve = 3 // prompt + window summary + footer
	const fallback = 12
	if m.height <= 0 {
		return fallback
	}
	if r := m.height - reserve; r >= 1 {
		return r
	}
	return 1
}
