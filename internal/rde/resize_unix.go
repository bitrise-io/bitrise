//go:build unix

package rde

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// watchWindowResize forwards local terminal resize events to the remote PTY.
// On every SIGWINCH it reads the current size of fd and sends a WindowChange
// request so the remote program reflows. The returned stop func unregisters
// the signal handler and ends the watcher goroutine.
func watchWindowResize(session *ssh.Session, fd int) (stop func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ch:
				w, h, err := term.GetSize(fd)
				if err != nil {
					continue
				}
				_ = session.WindowChange(h, w)
			}
		}
	}()

	return func() {
		signal.Stop(ch)
		close(done)
	}
}
