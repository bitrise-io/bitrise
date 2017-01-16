package colorstring

import (
	"fmt"
)

// Color ...
// ANSI color escape sequences
type Color string

const (
	blackColor   Color = "\x1b[30;1m"
	redColor     Color = "\x1b[31;1m"
	greenColor   Color = "\x1b[32;1m"
	yellowColor  Color = "\x1b[33;1m"
	blueColor    Color = "\x1b[34;1m"
	magentaColor Color = "\x1b[35;1m"
	cyanColor    Color = "\x1b[36;1m"
	resetColor   Color = "\x1b[0m"
)

// ColorFunc ...
type ColorFunc func(a ...interface{}) string

func addColor(color Color, msg string) string {
	return string(color) + msg + string(resetColor)
}

// NoColor ...
func NoColor(a ...interface{}) string {
	return fmt.Sprint(a...)
}

// Black ...
func Black(a ...interface{}) string {
	return addColor(blackColor, fmt.Sprint(a...))
}

// Red ...
func Red(a ...interface{}) string {
	return addColor(redColor, fmt.Sprint(a...))
}

// Green ...
func Green(a ...interface{}) string {
	return addColor(greenColor, fmt.Sprint(a...))
}

// Yellow ...
func Yellow(a ...interface{}) string {
	return addColor(yellowColor, fmt.Sprint(a...))
}

// Blue ...
func Blue(a ...interface{}) string {
	return addColor(blueColor, fmt.Sprint(a...))
}

// Magenta ...
func Magenta(a ...interface{}) string {
	return addColor(magentaColor, fmt.Sprint(a...))
}

// Cyan ...
func Cyan(a ...interface{}) string {
	return addColor(cyanColor, fmt.Sprint(a...))
}

// ColorfFunc ...
type ColorfFunc func(format string, a ...interface{}) string

// NoColorf ...
func NoColorf(format string, a ...interface{}) string {
	return NoColor(fmt.Sprintf(format, a...))
}

// Blackf ...
func Blackf(format string, a ...interface{}) string {
	return Black(fmt.Sprintf(format, a...))
}

// Redf ...
func Redf(format string, a ...interface{}) string {
	return Red(fmt.Sprintf(format, a...))
}

// Greenf ...
func Greenf(format string, a ...interface{}) string {
	return Green(fmt.Sprintf(format, a...))
}

// Yellowf ...
func Yellowf(format string, a ...interface{}) string {
	return Yellow(fmt.Sprintf(format, a...))
}

// Bluef ...
func Bluef(format string, a ...interface{}) string {
	return Blue(fmt.Sprintf(format, a...))
}

// Magentaf ...
func Magentaf(format string, a ...interface{}) string {
	return Magenta(fmt.Sprintf(format, a...))
}

// Cyanf ...
func Cyanf(format string, a ...interface{}) string {
	return Cyan(fmt.Sprintf(format, a...))
}
