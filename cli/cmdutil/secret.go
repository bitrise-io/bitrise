package cmdutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// terminalFd reports whether stream is an *os.File backed by a TTY. fd is
// only meaningful when isTerminal is true.
func terminalFd(stream any) (fd int, isTerminal bool) {
	f, ok := stream.(*os.File)
	if !ok {
		return 0, false
	}
	fd = int(f.Fd()) // file descriptors are small ints, no overflow risk
	return fd, term.IsTerminal(fd)
}

// IsTerminal reports whether r is an interactive terminal. Pipes and buffers
// never are, so callers can pick an interactive default (e.g. browser login)
// while keeping non-interactive stdin (CI, pipes) working.
func IsTerminal(r io.Reader) bool {
	_, ok := terminalFd(r)
	return ok
}

// ReadTokenInput reads a token from in, trimming all surrounding whitespace
// (tokens are forgiving of shell copy/paste artifacts). When fromStdin is
// true, or in isn't a terminal, it reads a line directly; otherwise it
// prompts and reads a masked line.
func ReadTokenInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool) (string, error) {
	return readSecretInput(in, stderr, prompt, fromStdin, strings.TrimSpace)
}

// ReadPasswordInput reads a password from in the same way as ReadTokenInput,
// but trims only a trailing line terminator — a leading/trailing space can
// be a deliberate part of a password and must not be silently stripped.
func ReadPasswordInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool) (string, error) {
	return readSecretInput(in, stderr, prompt, fromStdin, func(s string) string {
		return strings.TrimRight(s, "\r\n")
	})
}

func readSecretInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool, trim func(string) string) (string, error) {
	if fd, ok := terminalFd(in); ok && !fromStdin {
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
