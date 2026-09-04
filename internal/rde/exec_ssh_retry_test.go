package rde

import (
	"errors"
	"io"
	"net"
	"testing"
)

func TestIsRetryableDialErr(t *testing.T) {
	connRefused := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connect: connection refused")}
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"connection refused (net.OpError)", connRefused, true},
		{"wrapped net.OpError", errors.New("ssh dial host:port: " + connRefused.Error()), false}, // string-wrapped, type lost
		{"fmt-wrapped net.OpError", fmtWrap(connRefused), true},
		{"early EOF", io.EOF, true},
		{"auth failure (plain error)", errors.New("ssh: handshake failed: unable to authenticate"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableDialErr(tc.err); got != tc.want {
				t.Errorf("isRetryableDialErr(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// fmtWrap mimics dialSSH wrapping a net error with %w so errors.As still finds it.
func fmtWrap(err error) error {
	return errWrap{err}
}

type errWrap struct{ err error }

func (e errWrap) Error() string { return "ssh dial host:port: " + e.err.Error() }
func (e errWrap) Unwrap() error { return e.err }
