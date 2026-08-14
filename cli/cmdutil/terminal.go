package cmdutil

import (
	"io"
	"os"

	"golang.org/x/term"
)

// TerminalFd reports whether stream is an *os.File backed by a TTY. fd is
// only meaningful when isTerminal is true.
func TerminalFd(stream any) (fd int, isTerminal bool) {
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
	_, ok := TerminalFd(r)
	return ok
}

// IsTerminalWriter reports whether w is an interactive terminal, e.g. for
// gating pretty-printing so piped or redirected output isn't reformatted.
func IsTerminalWriter(w io.Writer) bool {
	_, ok := TerminalFd(w)
	return ok
}
