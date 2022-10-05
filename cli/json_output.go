package cli

import (
	"fmt"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"io"
	"os"
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
		log.Printf("failed to print message: %s, error: %s\n", f.String(), err)
	}
}

// JSONLoger ...
type JSONLoger struct {
	writer io.Writer
}

// NewJSONLoger ...
func NewJSONLoger(writer io.Writer) *JSONLoger {
	return &JSONLoger{
		writer: writer,
	}
}

// NewDefaultJSONLoger ...
func NewDefaultJSONLoger() JSONLoger {
	return JSONLoger{
		writer: os.Stdout,
	}
}

// Print ...
func (l JSONLoger) Print(f Formatable) {
	if _, err := fmt.Fprint(l.writer, f.JSON()); err != nil {
		log.Printf("failed to print message: %s, error: %s\n", f.JSON(), err)
	}
}
