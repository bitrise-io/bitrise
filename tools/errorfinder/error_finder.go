package errorfinder

import (
	"io"
)

// ErrorFinder wraps one or multiple standard `io.Writer` instances with `errorFindingWriter` which parses the stream
// coming via the `Write` method and keeps the latest "red" block (that matches \x1b[31;1m control sequence). `WrapError`
// method wraps an error with the latest seen "red" error message if any is present.
type ErrorFinder interface {
	WrapWriter(writer io.Writer) io.Writer
	WrapError(err error) error
}

type errorFinder struct {
	writers []errorFindingWriter
}

// NewErrorFinder ...
func NewErrorFinder() ErrorFinder {
	return &errorFinder{}
}

// WrapWriter ...
func (e *errorFinder) WrapWriter(writer io.Writer) io.Writer {
	result := newWriter(writer)
	e.writers = append(e.writers, result)
	return result
}

// WrapError ...
func (e *errorFinder) WrapError(err error) error {
	if err == nil {
		return nil
	}
	var ts int64
	var message string
	for _, writer := range e.writers {
		if msg := writer.getErrorMessage(); msg != nil && msg.Timestamp > ts {
			message = msg.Message
			ts = msg.Timestamp
		}
	}
	if message != "" {
		return &StepError{
			Message: message,
			Err:     err,
		}
	}
	return err
}
