package cmdutil

import (
	"bytes"
	"strings"
	"testing"
)

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

func TestReadSecretInput_StillFullyTrims(t *testing.T) {
	in := strings.NewReader("  tok  \n")
	got, err := ReadSecretInput(in, &bytes.Buffer{}, "", false)
	if err != nil {
		t.Fatalf("ReadSecretInput: %v", err)
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

func TestCheckPasswordStdinPiped_ErrorsWhenTerminal(t *testing.T) {
	err := CheckPasswordStdinPiped(true, true, "bitrise auth login --email <email>")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--password-stdin requires piped stdin") {
		t.Fatalf("error %q does not mention piped stdin", err.Error())
	}
	if !strings.Contains(err.Error(), "bitrise auth login --email <email> --password-stdin") {
		t.Fatalf("error %q does not include the example command", err.Error())
	}
}

func TestCheckValueStdinPiped(t *testing.T) {
	err := CheckValueStdinPiped(true, true, "bitrise rde saved-input create --key <key>")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--value-stdin requires piped stdin") {
		t.Fatalf("error %q does not mention piped stdin", err.Error())
	}
	if !strings.Contains(err.Error(), "bitrise rde saved-input create --key <key> --value-stdin") {
		t.Fatalf("error %q does not include the example command", err.Error())
	}
	if err := CheckValueStdinPiped(true, false, "x"); err != nil {
		t.Errorf("--value-stdin + piped input: %v", err)
	}
	if err := CheckValueStdinPiped(false, true, "x"); err != nil {
		t.Errorf("no --value-stdin, terminal: %v", err)
	}
}

func TestCheckPasswordStdinPiped_OKCases(t *testing.T) {
	const example = "bitrise user create --email <email>"
	if err := CheckPasswordStdinPiped(true, false, example); err != nil {
		t.Errorf("--password-stdin + piped input: %v", err)
	}
	if err := CheckPasswordStdinPiped(false, true, example); err != nil {
		t.Errorf("no --password-stdin, terminal: %v", err)
	}
	if err := CheckPasswordStdinPiped(false, false, example); err != nil {
		t.Errorf("no --password-stdin, non-terminal: %v", err)
	}
}
