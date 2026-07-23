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

// ReadSecretInput reads a secret (token, password) from in. When fromStdin is
// true, or in isn't a terminal, it reads a line directly; otherwise it prompts
// and reads a masked line.
func ReadSecretInput(in io.Reader, stderr io.Writer, prompt string, fromStdin bool) (string, error) {
	if fd, ok := terminalFd(in); ok && !fromStdin {
		if _, err := fmt.Fprint(stderr, prompt); err != nil {
			return "", err
		}
		b, err := term.ReadPassword(fd)
		if _, perr := fmt.Fprintln(stderr); perr != nil { // newline after no-echo input
			return "", perr
		}
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	s, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(s), nil
}
