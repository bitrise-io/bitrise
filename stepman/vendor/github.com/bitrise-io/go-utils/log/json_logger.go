package log

import (
	"fmt"
	"io"
	"os"
)

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
	fmt.Fprint(l.writer, f.JSON())
}
