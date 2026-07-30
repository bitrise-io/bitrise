package style

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew_NonTTYWriterIsAnsiFree(t *testing.T) {
	// A *bytes.Buffer is never a TTY → lipgloss/termenv falls back to the
	// Ascii profile and styles render as plain strings. This is the
	// invariant that keeps tests, pipes, and JSON output ANSI-free.
	var buf bytes.Buffer
	s := New(&buf)
	for _, in := range []string{"hello", "Build:", "success"} {
		if got := s.Header.Render(in); got != in {
			t.Errorf("Header.Render(%q) = %q, want plain", in, got)
		}
		if got := s.Success.Render(in); got != in {
			t.Errorf("Success.Render(%q) = %q, want plain", in, got)
		}
	}
}

func TestTable_HeadersAndRows(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)

	headers := []string{"NUMBER", "STATUS", "BRANCH"}
	rows := [][]string{
		{"42", "success", "main"},
		{"41", "in-progress", "feature/x"},
	}
	if err := Table(&buf, headers, rows, s.Header, nil); err != nil {
		t.Fatalf("Table: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NUMBER") || !strings.Contains(out, "STATUS") || !strings.Contains(out, "BRANCH") {
		t.Errorf("missing headers in output:\n%s", out)
	}
	if !strings.Contains(out, "feature/x") {
		t.Errorf("missing row content:\n%s", out)
	}
	// Three lines: header + 2 data.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
}

func TestTable_ColumnAlignment(t *testing.T) {
	// Cells of varying length should be right-padded so the next column
	// always starts at the same offset.
	var buf bytes.Buffer
	s := New(&buf)

	headers := []string{"A", "B"}
	rows := [][]string{
		{"x", "1"},
		{"longer-cell", "2"},
	}
	if err := Table(&buf, headers, rows, s.Header, nil); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines: %q", len(lines), buf.String())
	}
	col := strings.Index(lines[1], "1")
	if col == -1 || col != strings.Index(lines[2], "2") {
		t.Errorf("columns misaligned:\n%s", buf.String())
	}
}

func TestTable_StylerIsCalled(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)

	headers := []string{"X"}
	rows := [][]string{{"a"}, {"b"}}

	called := 0
	styler := func(_, _ int, content string) string {
		called++
		return "<" + content + ">"
	}
	if err := Table(&buf, headers, rows, s.Header, styler); err != nil {
		t.Fatal(err)
	}
	if called != 2 {
		t.Errorf("styler called %d times, want 2", called)
	}
	out := buf.String()
	if !strings.Contains(out, "<a>") || !strings.Contains(out, "<b>") {
		t.Errorf("styler output not in result:\n%s", out)
	}
}

func TestTable_FewerCellsThanHeadersDoesntPanic(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	headers := []string{"A", "B", "C"}
	rows := [][]string{
		{"only-a"}, // 1 cell for 3 headers
		{"a", "b", "c"},
	}
	if err := Table(&buf, headers, rows, s.Header, nil); err != nil {
		t.Fatalf("Table: %v", err)
	}
	if !strings.Contains(buf.String(), "only-a") {
		t.Errorf("first row missing:\n%s", buf.String())
	}
}

func TestTable_EmptyHeadersIsNoop(t *testing.T) {
	var buf bytes.Buffer
	s := New(&buf)
	if err := Table(&buf, nil, nil, s.Header, nil); err != nil {
		t.Fatalf("Table: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty headers, got %q", buf.String())
	}
}
