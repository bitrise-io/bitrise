package log

import (
	"fmt"
	"io"
	"os"
)

// RawLogger ...
type RawLogger struct {
	writer io.Writer
}

// NewRawLogger ...
func NewRawLogger(writer io.Writer) *RawLogger {
	return &RawLogger{
		writer: writer,
	}
}

// NewDefaultRawLogger ...
func NewDefaultRawLogger() RawLogger {
	return RawLogger{
		writer: os.Stdout,
	}
}

// Print ...
func (l RawLogger) Print(f Formatable) {
	if _, err := fmt.Fprintln(l.writer, f.String()); err != nil {
		fmt.Printf("failed to print message: %s, error: %s\n", f.String(), err)
	}
}
