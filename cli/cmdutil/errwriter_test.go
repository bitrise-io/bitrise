package cmdutil

import (
	"bytes"
	"errors"
	"testing"
)

func TestErrWriter_FAndLn(t *testing.T) {
	var buf bytes.Buffer
	ew := NewErrWriter(&buf)
	ew.F("count: %d", 3)
	ew.Ln("a", "b")
	if ew.Err != nil {
		t.Fatalf("unexpected error: %v", ew.Err)
	}
	want := "count: 3a b\n"
	if buf.String() != want {
		t.Errorf("buf = %q, want %q", buf.String(), want)
	}
}

func TestErrWriter_StopsAfterFirstError(t *testing.T) {
	w := &failAfterWriter{failOn: 2}
	ew := NewErrWriter(w)

	ew.Ln("first") // write 1: succeeds
	if ew.Err != nil {
		t.Fatalf("unexpected error after first write: %v", ew.Err)
	}
	ew.Ln("second") // write 2: fails
	firstErr := ew.Err
	if firstErr == nil {
		t.Fatal("expected an error after the second write")
	}
	ew.Ln("third") // must be a no-op
	ew.F("fourth")
	if ew.Err != firstErr {
		t.Errorf("Err changed after failure: got %v, want %v", ew.Err, firstErr)
	}
	if w.calls != 2 {
		t.Errorf("writer called %d times, want 2 (chain should stop after the first error)", w.calls)
	}
}

// failAfterWriter fails starting from its failOn-th Write call (1-indexed).
type failAfterWriter struct {
	calls  int
	failOn int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls >= w.failOn {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}
