package picker

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// items builds n simple rows titled a, b, c, … for navigation tests.
func threeItems() []Item {
	return []Item{{Title: "alpha"}, {Title: "beta"}, {Title: "gamma"}}
}

// drive applies one message and returns the resulting model.
func drive(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func keyType(tt tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: tt} }
func runes(s string) tea.KeyMsg         { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

// isQuit reports whether cmd is tea.Quit.
func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func newTestModel(cfg Config) model {
	if cfg.Out == nil {
		cfg.Out = &bytes.Buffer{}
	}
	if cfg.Items == nil {
		cfg.Items = threeItems()
	}
	return newModel(cfg)
}

func TestCursorStartsOnConfiguredIndex(t *testing.T) {
	if m := newTestModel(Config{Cursor: 1}); m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
	// Out-of-range cursors clamp to 0.
	if m := newTestModel(Config{Cursor: 99}); m.cursor != 0 {
		t.Errorf("oob cursor = %d, want 0", m.cursor)
	}
	if m := newTestModel(Config{Cursor: -3}); m.cursor != 0 {
		t.Errorf("negative cursor = %d, want 0", m.cursor)
	}
	// DefaultIdx out of range disables the badge (-1).
	if m := newTestModel(Config{DefaultIdx: 99}); m.defaultIdx != -1 {
		t.Errorf("oob defaultIdx = %d, want -1", m.defaultIdx)
	}
}

func TestNavigationClamps(t *testing.T) {
	m := newTestModel(Config{})
	m = drive(m, keyType(tea.KeyDown))
	m = drive(m, keyType(tea.KeyDown))
	if m.cursor != 2 {
		t.Fatalf("after 2×down cursor = %d, want 2", m.cursor)
	}
	// Can't move past the last row.
	m = drive(m, keyType(tea.KeyDown))
	if m.cursor != 2 {
		t.Fatalf("down past end cursor = %d, want 2 (clamped)", m.cursor)
	}
	// j/k mirror the arrows when filtering is off.
	m = drive(m, runes("k"))
	if m.cursor != 1 {
		t.Fatalf("after k cursor = %d, want 1", m.cursor)
	}
	m = drive(m, keyType(tea.KeyUp))
	m = drive(m, keyType(tea.KeyUp))
	if m.cursor != 0 {
		t.Fatalf("up past start cursor = %d, want 0 (clamped)", m.cursor)
	}
}

func TestEnterConfirmsAndQuits(t *testing.T) {
	m := newTestModel(Config{Cursor: 1})
	next, cmd := m.Update(keyType(tea.KeyEnter))
	m = next.(model)
	if m.chosen != 1 {
		t.Errorf("chosen = %d, want 1", m.chosen)
	}
	if !isQuit(cmd) {
		t.Error("expected Enter to issue tea.Quit")
	}
}

func TestCancelKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"esc", keyType(tea.KeyEsc)},
		{"ctrl-c", keyType(tea.KeyCtrlC)},
		{"q (filter off)", runes("q")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, cmd := newTestModel(Config{}).Update(tc.msg)
			m := next.(model)
			if !m.cancelled {
				t.Error("expected cancelled=true")
			}
			if m.chosen != -1 {
				t.Errorf("chosen = %d, want -1", m.chosen)
			}
			if !isQuit(cmd) {
				t.Error("expected tea.Quit")
			}
		})
	}
}

func TestFilterNarrowsAndRestores(t *testing.T) {
	m := newTestModel(Config{Filter: true})
	// "al" matches only "alpha".
	m = drive(m, runes("a"))
	m = drive(m, runes("l"))
	if len(m.view) != 1 || m.view[0] != 0 {
		t.Fatalf("view after 'al' = %v, want [0]", m.view)
	}
	if m.cursor != 0 {
		t.Fatalf("cursor after filter = %d, want 0", m.cursor)
	}
	// Enter selects the single match.
	next, cmd := m.Update(keyType(tea.KeyEnter))
	if got := next.(model).chosen; got != 0 {
		t.Fatalf("chosen = %d, want 0", got)
	}
	if !isQuit(cmd) {
		t.Error("expected tea.Quit on filtered Enter")
	}
	// Backspace widens the match again.
	m = drive(m, keyType(tea.KeyBackspace)) // → "a", matches all three
	if len(m.view) != 3 {
		t.Fatalf("view after backspace = %v, want all 3", m.view)
	}
}

func TestFilterEmptyResultIgnoresEnter(t *testing.T) {
	m := newTestModel(Config{Filter: true})
	m = drive(m, runes("z")) // matches nothing
	if len(m.view) != 0 {
		t.Fatalf("view = %v, want empty", m.view)
	}
	next, cmd := m.Update(keyType(tea.KeyEnter))
	if got := next.(model).chosen; got != -1 {
		t.Errorf("chosen = %d, want -1 (no-op on empty)", got)
	}
	if isQuit(cmd) {
		t.Error("Enter on empty result should not quit")
	}
}

func TestQTypesIntoFilterWhenFilterOn(t *testing.T) {
	m := newTestModel(Config{Filter: true})
	m = drive(m, runes("q"))
	if m.cancelled {
		t.Error("q should type into the filter, not cancel, when Filter is on")
	}
	if m.filter != "q" {
		t.Errorf("filter = %q, want %q", m.filter, "q")
	}
}

func TestViewContentAndCollapse(t *testing.T) {
	m := newTestModel(Config{Prompt: "Select an image", Cursor: 1, DefaultIdx: 1})
	v := m.View()
	for _, want := range []string{"Select an image", "›", "(default)", "alpha", "beta"} {
		if !strings.Contains(v, want) {
			t.Errorf("View() missing %q in:\n%s", want, v)
		}
	}
	// After confirming, the frame collapses so the caller's line is what shows.
	m = drive(m, keyType(tea.KeyEnter))
	if got := m.View(); got != "" {
		t.Errorf("View() after confirm = %q, want empty", got)
	}
	// Same after cancelling.
	c := drive(newTestModel(Config{}), keyType(tea.KeyEsc))
	if got := c.View(); got != "" {
		t.Errorf("View() after cancel = %q, want empty", got)
	}
}

func TestViewRendersNote(t *testing.T) {
	m := newTestModel(Config{
		Prompt: "Last used for this project",
		Note:   "  Image         Ubuntu 24.04\n  Machine type  g2.mac",
		Items:  []Item{{Title: "Use this setup"}, {Title: "Change image"}},
	})
	v := m.View()
	for _, want := range []string{"Image         Ubuntu 24.04", "Machine type  g2.mac", "Use this setup"} {
		if !strings.Contains(v, want) {
			t.Errorf("View() missing note/row %q in:\n%s", want, v)
		}
	}
}

func TestStatusRendersInView(t *testing.T) {
	m := newTestModel(Config{Items: []Item{
		{Title: "sess-1", Desc: "[main] 2h ago", Status: "running", Tone: ToneSuccess},
	}})
	v := m.View()
	for _, want := range []string{"sess-1", "[main] 2h ago", "(running)"} {
		if !strings.Contains(v, want) {
			t.Errorf("View() missing %q in:\n%s", want, v)
		}
	}
}

func TestToneStyleMapping(t *testing.T) {
	m := newTestModel(Config{})
	for _, tc := range []struct {
		tone Tone
		want lipgloss.Style
	}{
		{ToneSuccess, m.s.Success},
		{ToneWarn, m.s.Warn},
		{ToneDanger, m.s.Failure},
		{ToneDim, m.s.Dim},
		{ToneNone, m.s.Dim},
	} {
		if !reflect.DeepEqual(m.toneStyle(tc.tone), tc.want) {
			t.Errorf("toneStyle(%v) mapped to the wrong style", tc.tone)
		}
	}
	// Distinct tones must map to distinct styles, else coloring is meaningless.
	if reflect.DeepEqual(m.s.Success, m.s.Failure) {
		t.Error("Success and Failure styles should differ")
	}
}

func TestViewportWindowsLongList(t *testing.T) {
	items := make([]Item, 30)
	for i := range items {
		items[i] = Item{Title: string(rune('a' + i%26))}
	}
	m := newTestModel(Config{Items: items})
	m = drive(m, tea.WindowSizeMsg{Width: 80, Height: 10}) // visibleRows = 10-3 = 7
	if got := m.visibleRows(); got != 7 {
		t.Fatalf("visibleRows = %d, want 7", got)
	}
	for i := 0; i < 10; i++ {
		m = drive(m, keyType(tea.KeyDown))
	}
	if m.cursor != 10 {
		t.Fatalf("cursor = %d, want 10", m.cursor)
	}
	// Window scrolled to keep the cursor visible: offset = cursor-maxRows+1.
	if m.offset != 4 {
		t.Fatalf("offset = %d, want 4", m.offset)
	}
	if v := m.View(); !strings.Contains(v, "of 30") {
		t.Errorf("View() should show the windowed summary, got:\n%s", v)
	}
}

// dividerItems builds a grouped list: [G1 divider, alpha, beta, G2 divider, gamma].
func dividerItems() []Item {
	return []Item{
		{Title: "G1", Divider: true},
		{Title: "alpha"},
		{Title: "beta"},
		{Title: "G2", Divider: true},
		{Title: "gamma"},
	}
}

func TestCursorOpensOffDivider(t *testing.T) {
	// Cursor 0 is a divider → it advances to the first selectable row (index 1).
	if m := newTestModel(Config{Items: dividerItems(), Cursor: 0}); m.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (first selectable, not the divider)", m.cursor)
	}
	// A cursor already on a selectable row is left alone.
	if m := newTestModel(Config{Items: dividerItems(), Cursor: 2}); m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}
}

func TestNavigationSkipsDividers(t *testing.T) {
	m := newTestModel(Config{Items: dividerItems()}) // opens on 1 (alpha)
	m = drive(m, keyType(tea.KeyDown))               // → 2 (beta)
	if m.cursor != 2 {
		t.Fatalf("after down cursor = %d, want 2 (beta)", m.cursor)
	}
	m = drive(m, keyType(tea.KeyDown)) // skips the G2 divider → 4 (gamma)
	if m.cursor != 4 {
		t.Fatalf("after down cursor = %d, want 4 (gamma, divider skipped)", m.cursor)
	}
	m = drive(m, keyType(tea.KeyDown)) // at the last row: stays put
	if m.cursor != 4 {
		t.Fatalf("down past end cursor = %d, want 4", m.cursor)
	}
	m = drive(m, keyType(tea.KeyUp)) // skips G2 back up → 2 (beta)
	if m.cursor != 2 {
		t.Fatalf("after up cursor = %d, want 2 (beta)", m.cursor)
	}
	m = drive(m, keyType(tea.KeyUp)) // → 1 (alpha)
	m = drive(m, keyType(tea.KeyUp)) // top selectable; G1 divider above → stays at 1
	if m.cursor != 1 {
		t.Fatalf("up past top cursor = %d, want 1 (alpha)", m.cursor)
	}
}

func TestEnterIgnoresDivider(t *testing.T) {
	// Force the cursor onto a divider (navigation never does this) and confirm
	// Enter is a no-op — the frame stays open and nothing is chosen.
	m := newTestModel(Config{Items: dividerItems()})
	m.cursor = 0 // the G1 divider
	next, cmd := m.Update(keyType(tea.KeyEnter))
	got := next.(model)
	if got.chosen != -1 {
		t.Errorf("chosen = %d, want -1 (divider not selectable)", got.chosen)
	}
	if isQuit(cmd) {
		t.Error("Enter on a divider should not quit")
	}
}

func TestDividerRendersAsHeaderWithoutCursor(t *testing.T) {
	m := newTestModel(Config{Items: dividerItems()}) // cursor on alpha (index 1)
	lines := strings.Split(m.View(), "\n")
	var g1 string
	for _, ln := range lines {
		if strings.Contains(ln, "G1") {
			g1 = ln
		}
	}
	if g1 == "" {
		t.Fatalf("View() missing the G1 divider:\n%s", m.View())
	}
	if strings.Contains(g1, "›") {
		t.Errorf("divider line should not carry the cursor marker: %q", g1)
	}
}

func TestSelectEmptyItemsCancels(t *testing.T) {
	_, err := Select(t.Context(), Config{Items: nil, Out: &bytes.Buffer{}})
	if err != ErrCancelled {
		t.Errorf("Select(empty) err = %v, want ErrCancelled", err)
	}
}
