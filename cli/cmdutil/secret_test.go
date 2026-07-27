package cmdutil

import (
	"bytes"
	"os"
	"strings"
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
	tmp, err := os.CreateTemp(t.TempDir(), "tty-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tmp.Close() })
	if IsTerminal(tmp) {
		t.Error("IsTerminal(regular *os.File) should be false")
	}
}

func TestReadSecretInput_NonTerminalReadsLine(t *testing.T) {
	in := strings.NewReader("  a-token-value  \nrest\n")
	var stderr bytes.Buffer

	got, err := ReadSecretInput(in, &stderr, "Token: ", false)
	if err != nil {
		t.Fatalf("ReadSecretInput: %v", err)
	}
	if got != "a-token-value" {
		t.Fatalf("got %q, want %q", got, "a-token-value")
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no prompt written for a non-terminal reader, got %q", stderr.String())
	}
}

func TestReadSecretInput_EOFWithoutNewline(t *testing.T) {
	in := strings.NewReader("no-trailing-newline")
	got, err := ReadSecretInput(in, &bytes.Buffer{}, "", true)
	if err != nil {
		t.Fatalf("ReadSecretInput: %v", err)
	}
	if got != "no-trailing-newline" {
		t.Fatalf("got %q", got)
	}
}
