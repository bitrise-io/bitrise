package corelog

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ConsoleTimeLayout ...
const ConsoleTimeLayout = "15:04:05"

var levelToANSIColorCode = map[Level]ANSIColorCode{
	ErrorLevel: RedCode,
	WarnLevel:  YellowCode,
	InfoLevel:  BlueCode,
	DoneLevel:  GreenCode,
	DebugLevel: MagentaCode,
}

type ANSIColorCode string

const (
	RedCode     ANSIColorCode = "\x1b[31;1m"
	YellowCode  ANSIColorCode = "\x1b[33;1m"
	BlueCode    ANSIColorCode = "\x1b[34;1m"
	GreenCode   ANSIColorCode = "\x1b[32;1m"
	MagentaCode ANSIColorCode = "\x1b[35;1m"
	ResetCode   ANSIColorCode = "\x1b[0m"
)

type consoleLogger struct {
	output io.Writer
}

func newConsoleLogger(output io.Writer) *consoleLogger {
	return &consoleLogger{
		output: output,
	}

}

// LogMessage ...
func (l *consoleLogger) LogMessage(message string, fields MessageFields) {
	message = addColor(fields.Level, message)

	var prefixes []string
	if fields.Timestamp != "" {
		prefixes = append(prefixes, fmt.Sprintf("[%s]", fields.Timestamp))
	}
	if fields.Producer != "" {
		prefixes = append(prefixes, string(fields.Producer))
	}
	if fields.ProducerID != "" {
		prefixes = append(prefixes, fields.ProducerID)
	}
	prefix := strings.Join(prefixes, " ")
	if prefix != "" {
		prefix += " "
	}

	message = prefix + message
	if _, err := fmt.Fprint(l.output, message); err != nil {
		// Encountered an error during writing the message to the output. Manually construct a message for
		// the error and print it to the stdout.
		fmt.Printf("writing log message failed: %s", err)
	}
}

func CreateConsoleLogMessageFields(level Level, timeProvider func() time.Time) MessageFields {
	fields := MessageFields{
		Level: level,
	}
	if timeProvider != nil {
		fields.Timestamp = timeProvider().Format(ConsoleTimeLayout)
	}
	return fields
}

func addColor(level Level, message string) string {
	color := levelToANSIColorCode[level]
	if color != "" {
		return string(color) + message + string(ResetCode)
	}
	return message
}