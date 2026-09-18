package cmdutil

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
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

// YAMLLogger ...
type YAMLLogger struct {
	writer io.Writer
}

// NewDefaultYAMLLogger ...
func NewDefaultYAMLLogger() YAMLLogger {
	return YAMLLogger{
		writer: os.Stdout,
	}
}

// Print marshals the concrete model behind f rather than asking it for a
// rendering: Formatable has no YAML method, and only the two loggers above
// need a hand-written one. Models reaching here need yaml tags matching their
// json tags, like anything else passed to output.Print.
func (l YAMLLogger) Print(f Formatable) {
	if err := output.Print(l.writer, f, output.FormatYML); err != nil {
		log.Printf("failed to print message: %s, error: %s\n", f.String(), err)
	}
}
