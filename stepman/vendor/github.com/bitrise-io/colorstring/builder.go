package colorstring

import (
	"fmt"
)

// Builder is a struct to build multiline colored string.
type Builder struct {
	s string
}

// NewBuilder creates a new Builder instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// add appends the given formatted string colored by the given color function to the Builder.
func (b *Builder) add(colorFunc ColorfFunc, format string, v ...interface{}) *Builder {
	if colorFunc == nil {
		b.s += fmt.Sprintf(format, v...)
	} else {
		b.s += colorFunc(format, v...)
	}
	return b
}

// NewLine appends a newline to the Builder.
func (b *Builder) NewLine() *Builder {
	return b.add(nil, "\n")
}

// Plain appends the given formatted string to the Builder.
func (b *Builder) Plain(format string, v ...interface{}) *Builder {
	return b.add(nil, format, v...)
}

// Black appends the given formatted, black colored string to the Builder.
func (b *Builder) Black(format string, v ...interface{}) *Builder {
	return b.add(Black, format, v...)
}

// Red appends the given formatted, red colored string to the Builder.
func (b *Builder) Red(format string, v ...interface{}) *Builder {
	return b.add(Red, format, v...)
}

// Green appends the given formatted, green colored string to the Builder.
func (b *Builder) Green(format string, v ...interface{}) *Builder {
	return b.add(Green, format, v...)
}

// Yellow appends the given formatted, yellow colored string to the Builder.
func (b *Builder) Yellow(format string, v ...interface{}) *Builder {
	return b.add(Yellow, format, v...)
}

// Blue appends the given formatted, blue colored string to the Builder.
func (b *Builder) Blue(format string, v ...interface{}) *Builder {
	return b.add(Blue, format, v...)
}

// Magenta appends the given formatted, magenta colored string to the Builder.
func (b *Builder) Magenta(format string, v ...interface{}) *Builder {
	return b.add(Magenta, format, v...)
}

// Cyan appends the given formatted, cyan colored string to the Builder.
func (b *Builder) Cyan(format string, v ...interface{}) *Builder {
	return b.add(Cyan, format, v...)
}

// String returns the generated string.
func (b Builder) String() string {
	return b.s
}
