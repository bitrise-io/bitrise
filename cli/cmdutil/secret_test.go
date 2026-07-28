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

func TestReadTokenInput_NonTerminalReadsLine(t *testing.T) {
	in := strings.NewReader("  a-token-value  \nrest\n")
	var stderr bytes.Buffer

	got, err := ReadTokenInput(in, &stderr, "Token: ", false)
	if err != nil {
		t.Fatalf("ReadTokenInput: %v", err)
	}
	if got != "a-token-value" {
		t.Fatalf("got %q, want %q", got, "a-token-value")
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no prompt written for a non-terminal reader, got %q", stderr.String())
	}
}

func TestReadTokenInput_EOFWithoutNewline(t *testing.T) {
	in := strings.NewReader("no-trailing-newline")
	got, err := ReadTokenInput(in, &bytes.Buffer{}, "", true)
	if err != nil {
		t.Fatalf("ReadTokenInput: %v", err)
	}
	if got != "no-trailing-newline" {
		t.Fatalf("got %q", got)
	}
}

func TestReadTokenInput_StillFullyTrims(t *testing.T) {
	in := strings.NewReader("  tok  \n")
	got, err := ReadTokenInput(in, &bytes.Buffer{}, "", false)
	if err != nil {
		t.Fatalf("ReadTokenInput: %v", err)
	}
	if got != "tok" {
		t.Fatalf("got %q, want %q", got, "tok")
	}
}

func TestReadPasswordInput_PreservesInternalWhitespace(t *testing.T) {
	in := strings.NewReader(" hunter 2 \n")
	got, err := ReadPasswordInput(in, &bytes.Buffer{}, "", true)
	if err != nil {
		t.Fatalf("ReadPasswordInput: %v", err)
	}
	if got != " hunter 2 " {
		t.Fatalf("got %q, want %q (only trailing newline should be trimmed)", got, " hunter 2 ")
	}
}

func TestReadPasswordInput_StripsTrailingCRLF(t *testing.T) {
	in := strings.NewReader("hunter2\r\n")
	got, err := ReadPasswordInput(in, &bytes.Buffer{}, "", true)
	if err != nil {
		t.Fatalf("ReadPasswordInput: %v", err)
	}
	if got != "hunter2" {
		t.Fatalf("got %q, want %q", got, "hunter2")
	}
}

func TestReadPasswordInput_NoTrailingNewline(t *testing.T) {
	// Mirrors `printf '%s' "$PW" | ... --password-stdin` (no newline at all).
	in := strings.NewReader("hunter2")
	got, err := ReadPasswordInput(in, &bytes.Buffer{}, "", true)
	if err != nil {
		t.Fatalf("ReadPasswordInput: %v", err)
	}
	if got != "hunter2" {
		t.Fatalf("got %q, want %q", got, "hunter2")
	}
}
