package cmdutil

import (
	"bytes"
	"os"
	"testing"
)

func TestIsTerminal_NonTerminalInputs(t *testing.T) {
	if IsTerminal(nil) {
		t.Error("IsTerminal(nil) should be false")
	}
	if IsTerminal(&bytes.Buffer{}) {
		t.Error("IsTerminal(*bytes.Buffer) should be false")
	}
	// A regular file (not a TTY) is an *os.File, but term.IsTerminal is
	// false → IsTerminal should still return false.
	tmp := tempFile(t)
	if IsTerminal(tmp) {
		t.Error("IsTerminal(regular *os.File) should be false")
	}
}

func TestIsTerminalWriter_NonTerminalOutputs(t *testing.T) {
	if IsTerminalWriter(nil) {
		t.Error("IsTerminalWriter(nil) should be false")
	}
	if IsTerminalWriter(&bytes.Buffer{}) {
		t.Error("IsTerminalWriter(*bytes.Buffer) should be false")
	}
	if IsTerminalWriter(tempFile(t)) {
		t.Error("IsTerminalWriter(regular *os.File) should be false")
	}
}

func tempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "tty-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}
