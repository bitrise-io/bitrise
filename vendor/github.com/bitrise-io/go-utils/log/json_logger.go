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
	if _, err := fmt.Fprint(l.writer, f.JSON()); err != nil {
		fmt.Printf("failed to print message: %s, error: %s\n", f.JSON(), err)
	}
}
