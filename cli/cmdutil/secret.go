package cmdutil

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

// ReadSecretInput reads a secret from in, trimming all surrounding whitespace
// (tokens are forgiving of shell copy/paste artifacts). When fromStdin is
// true, or in isn't a terminal, it reads a line directly; otherwise it
// prompts and reads a masked line.
func ReadSecretInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool) (string, error) {
	return readSecretInput(in, stderr, prompt, fromStdin, strings.TrimSpace)
}

// ReadPasswordInput reads a password from in the same way as ReadSecretInput,
// but trims only a trailing line terminator — a leading/trailing space can
// be a deliberate part of a password and must not be silently stripped.
func ReadPasswordInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool) (string, error) {
	return readSecretInput(in, stderr, prompt, fromStdin, func(s string) string {
		return strings.TrimRight(s, "\r\n")
	})
}

// CheckPasswordStdinPiped reports an error when --password-stdin is set but
// stdin is an interactive terminal rather than a pipe. Falling through to a
// plain (non-masked) read in that case would block on terminal input with
// local echo still on, silently printing the password to the screen.
// exampleCmd is the invocation shown in the error's pipe example, e.g.
// "bitrise auth login --email <email>".
func CheckPasswordStdinPiped(passwordStdin, isTerminal bool, exampleCmd string) error {
	if passwordStdin && isTerminal {
		return fmt.Errorf("--password-stdin requires piped stdin (got an interactive terminal); pipe the password in, e.g.:\n  printf '%%s' \"$PW\" | %s --password-stdin", exampleCmd)
	}
	return nil
}

// CheckValueStdinPiped is CheckPasswordStdinPiped's counterpart for commands
// whose stdin secret arrives under --value-stdin (`rde saved-input`). Same
// hazard, same shape — only the flag and the prose differ, so they stay as two
// literal messages rather than one parameterized template.
func CheckValueStdinPiped(valueStdin, isTerminal bool, exampleCmd string) error {
	if valueStdin && isTerminal {
		return fmt.Errorf("--value-stdin requires piped stdin (got an interactive terminal); pipe the value in, e.g.:\n  printf '%%s' \"$VALUE\" | %s --value-stdin", exampleCmd)
	}
	return nil
}

func readSecretInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool, trim func(string) string) (string, error) {
	if fd, ok := TerminalFd(in); ok && !fromStdin {
		if _, err := fmt.Fprint(stderr, prompt); err != nil {
			return "", err
		}
		b, readErr := term.ReadPassword(fd)
		_, writeErr := fmt.Fprintln(stderr) // newline after no-echo input
		if readErr != nil {
			return "", readErr
		}
		if writeErr != nil {
			return "", writeErr
		}
		return trim(string(b)), nil
	}
	s, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return trim(s), nil
}
