//go:build !unix

package rde

import "golang.org/x/crypto/ssh"

// watchWindowResize is a no-op on platforms without SIGWINCH (e.g. Windows).
// The remote PTY keeps the size set at RequestPty time.
func watchWindowResize(_ *ssh.Session, _ int) (stop func()) {
	return func() {}
}
