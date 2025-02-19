package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
)

// Logger ...
type Logger interface {
	Print(f Formatable)
}

// Formatable ...
type Formatable interface {
	String() string
	JSON() string
}

// RawLogger ...
type RawLogger struct {
	writer io.Writer
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
		log.Printf("failed to print message: %s, error: %s\n", f.String(), err)
	}
}

// JSONLogger ...
type JSONLogger struct {
	writer io.Writer
}

// NewDefaultJSONLogger ...
func NewDefaultJSONLogger() JSONLogger {
	return JSONLogger{
		writer: os.Stdout,
	}
}

// Print ...
func (l JSONLogger) Print(f Formatable) {
	if _, err := fmt.Fprint(l.writer, f.JSON()); err != nil {
		log.Printf("failed to print message: %s, error: %s\n", f.JSON(), err)
	}
}
