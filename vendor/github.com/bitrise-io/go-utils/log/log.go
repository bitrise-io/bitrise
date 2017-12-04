package log

import (
	"fmt"
	"time"

	"io"
	"os"

	"github.com/bitrise-io/go-utils/colorstring"
)

//
// Log configuration

var timestampLayout = "15:04:05"

// SetTimestampLayout ...
func SetTimestampLayout(layout string) {
	timestampLayout = layout
}

var outWriter io.Writer = os.Stdout

// SetOutWriter ...
func SetOutWriter(writer io.Writer) {
	outWriter = writer
}

//
// Print with color

func printfWithColor(color colorstring.ColorfFunc, format string, v ...interface{}) {
	strWithColor := color(format, v...)
	fmt.Fprintln(outWriter, strWithColor)
}

// Printf ...
func Printf(format string, v ...interface{}) {
	printfWithColor(colorstring.NoColorf, format, v...)
}

// Infof ...
func Infof(format string, v ...interface{}) {
	printfWithColor(colorstring.Bluef, format, v...)
}

// Donef ...
func Donef(format string, v ...interface{}) {
	printfWithColor(colorstring.Greenf, format, v...)
}

// Errorf ...
func Errorf(format string, v ...interface{}) {
	printfWithColor(colorstring.Redf, format, v...)
}

// Warnf ...
func Warnf(format string, v ...interface{}) {
	printfWithColor(colorstring.Yellowf, format, v...)
}

//
// Print with color and timestamp

func timestamp() string {
	currentTime := time.Now()
	return currentTime.Format(timestampLayout)
}

func printfWithColorAndTime(color colorstring.ColorfFunc, format string, v ...interface{}) {
	strWithColor := color(format, v...)
	strWithColorAndTime := fmt.Sprintf("[%s] %s", timestamp(), strWithColor)
	fmt.Fprintln(outWriter, strWithColorAndTime)
}

// Printft ...
func Printft(format string, v ...interface{}) {
	printfWithColorAndTime(colorstring.NoColorf, format, v...)
}

// Infoft ...
func Infoft(format string, v ...interface{}) {
	printfWithColorAndTime(colorstring.Bluef, format, v...)
}

// Doneft ...
func Doneft(format string, v ...interface{}) {
	printfWithColorAndTime(colorstring.Greenf, format, v...)
}

// Errorft ...
func Errorft(format string, v ...interface{}) {
	printfWithColorAndTime(colorstring.Redf, format, v...)
}

// Warnft ...
func Warnft(format string, v ...interface{}) {
	printfWithColorAndTime(colorstring.Yellowf, format, v...)
}
