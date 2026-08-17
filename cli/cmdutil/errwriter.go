package cmdutil

import (
	"fmt"
	"io"
)

// ErrWriter wraps an io.Writer and captures the first write error so callers
// can chain writes and check once at the end.
type ErrWriter struct {
	w   io.Writer
	Err error
}

// NewErrWriter returns an ErrWriter backed by w.
func NewErrWriter(w io.Writer) *ErrWriter { return &ErrWriter{w: w} }

// F writes a formatted string, skipping if a previous write already failed.
func (ew *ErrWriter) F(format string, a ...any) {
	if ew.Err != nil {
		return
	}
	_, ew.Err = fmt.Fprintf(ew.w, format, a...)
}

// Ln writes args followed by a newline, skipping if a previous write failed.
func (ew *ErrWriter) Ln(a ...any) {
	if ew.Err != nil {
		return
	}
	_, ew.Err = fmt.Fprintln(ew.w, a...)
}
