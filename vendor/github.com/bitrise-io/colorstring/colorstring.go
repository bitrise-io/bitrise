package colorstring

import (
	"fmt"
)

// Color is the type of the ANSI color escape sequences.
type Color string

// ANSI color escape sequences.
const (
	blackColor   Color = "\x1b[30;1m"
	redColor     Color = "\x1b[31;1m"
	greenColor   Color = "\x1b[32;1m"
	yellowColor  Color = "\x1b[33;1m"
	blueColor    Color = "\x1b[34;1m"
	magentaColor Color = "\x1b[35;1m"
	cyanColor    Color = "\x1b[36;1m"
	reset        Color = "\x1b[0m"
)

// addColor colors the given string.
func addColor(c Color, s string) string {
	return string(c) + s + string(reset)
}

// ColorfFunc is the type of the specific color functions.
type ColorfFunc func(format string, a ...interface{}) string

// Plain returns the given formatted string.
func Plain(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

// Black adds black color to the given formatted string.
func Black(format string, a ...interface{}) string {
	return addColor(blackColor, fmt.Sprintf(format, a...))
}

// Red adds red color to the given formatted string.
func Red(format string, a ...interface{}) string {
	return addColor(redColor, fmt.Sprintf(format, a...))
}

// Green adds green color to the given formatted string.
func Green(format string, a ...interface{}) string {
	return addColor(greenColor, fmt.Sprintf(format, a...))
}

// Yellow adds yellow color to the given formatted string.
func Yellow(format string, a ...interface{}) string {
	return addColor(yellowColor, fmt.Sprintf(format, a...))
}

// Blue adds blue color to the given formatted string.
func Blue(format string, a ...interface{}) string {
	return addColor(blueColor, fmt.Sprintf(format, a...))
}

// Magenta adds magenta color to the given formatted string.
func Magenta(format string, a ...interface{}) string {
	return addColor(magentaColor, fmt.Sprintf(format, a...))
}

// Cyan adds cyan color to the given formatted string.
func Cyan(format string, a ...interface{}) string {
	return addColor(cyanColor, fmt.Sprintf(format, a...))
}
