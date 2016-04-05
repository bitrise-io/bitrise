package colorstring

import (
	"fmt"
)

const (
	clrBlack   = "\x1b[30;1m"
	clrRed     = "\x1b[31;1m"
	clrGreen   = "\x1b[32;1m"
	clrYellow  = "\x1b[33;1m"
	clrBlue    = "\x1b[34;1m"
	clrMagenta = "\x1b[35;1m"
	clrCyan    = "\x1b[36;1m"
	clrReset   = "\x1b[0m"
)

// Black ...
func Black(s string) string {
	return Blackf(s)
}

// Blackf ...
func Blackf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrBlack)
}

// Red ...
func Red(s string) string {
	return Redf(s)
}

// Redf ...
func Redf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrRed)
}

// Green ...
func Green(s string) string {
	return Greenf(s)
}

// Greenf ...
func Greenf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrGreen)
}

// Yellow ...
func Yellow(s string) string {
	return Yellowf(s)
}

// Yellowf ...
func Yellowf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrYellow)
}

// Blue ...
func Blue(s string) string {
	return Bluef(s)
}

// Bluef ...
func Bluef(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrBlue)
}

// Magenta ...
func Magenta(s string) string {
	return Magentaf(s)
}

// Magentaf ...
func Magentaf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrMagenta)
}

// Cyan ...
func Cyan(s string) string {
	return Cyanf(s)
}

// Cyanf ...
func Cyanf(format string, a ...interface{}) string {
	return addColor(fmt.Sprintf(format, a...), clrCyan)
}

func addColor(s, color string) string {
	return color + s + clrReset
}
