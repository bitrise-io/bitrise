package claude

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/style"
)

func TestSpinnerModel_ViewShowsLabelAndStatus(t *testing.T) {
	m := newSpinnerModel("Booting session…", style.New(&bytes.Buffer{}))

	view := m.View()
	if !strings.Contains(view, "Booting session…") {
		t.Errorf("view %q does not contain the label", view)
	}

	updated, _ := m.Update(statusMsg("starting"))
	m = updated.(spinnerModel)
	if !strings.Contains(m.View(), "(starting)") {
		t.Errorf("view %q does not show the live status", m.View())
	}
}

func TestSpinnerModel_CollapsesWhenDone(t *testing.T) {
	m := newSpinnerModel("Booting session…", style.New(&bytes.Buffer{}))

	updated, cmd := m.Update(doneMsg{})
	m = updated.(spinnerModel)
	if !m.finished {
		t.Error("model not marked finished after doneMsg")
	}
	if cmd == nil {
		t.Error("doneMsg should return a quit command")
	}
	if m.View() != "" {
		t.Errorf("finished view = %q, want empty (collapsed)", m.View())
	}
}

func TestAwait_FallbackRunsFnAndSkipsAnimation(t *testing.T) {
	// A bytes.Buffer is not a TTY, so await takes the plain-line fallback.
	var buf bytes.Buffer
	log := &stepLogger{w: &buf, s: style.New(&buf)}

	ran := false
	err := log.await(context.Background(), "Booting session…", "Session booted",
		func(_ context.Context, status func(string)) error {
			ran = true
			status("starting") // must be a safe no-op in the fallback
			return nil
		})
	if err != nil {
		t.Fatalf("await: %v", err)
	}
	if !ran {
		t.Error("fn was not run")
	}
	out := buf.String()
	if !strings.Contains(out, "Booting session…") {
		t.Errorf("output %q does not contain the label", out)
	}
	// The fallback keeps today's behavior: a plain step line, no settled "✓".
	if strings.Contains(out, "✓") {
		t.Errorf("output %q unexpectedly contains a settled check", out)
	}
}

func TestAwait_FallbackPropagatesError(t *testing.T) {
	var buf bytes.Buffer
	log := &stepLogger{w: &buf, s: style.New(&buf)}

	want := errors.New("boom")
	err := log.await(context.Background(), "Booting session…", "Session booted",
		func(_ context.Context, _ func(string)) error { return want })
	if !errors.Is(err, want) {
		t.Errorf("await error = %v, want %v", err, want)
	}
}

func TestAwait_QuietSuppressesOutput(t *testing.T) {
	var buf bytes.Buffer
	log := &stepLogger{w: &buf, s: style.New(&buf), quiet: true}

	err := log.await(context.Background(), "Booting session…", "Session booted",
		func(_ context.Context, _ func(string)) error { return nil })
	if err != nil {
		t.Fatalf("await: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("quiet await wrote %q, want nothing", buf.String())
	}
}
