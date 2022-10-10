package corelog

import (
	"fmt"
	"io"
)

var levelToANSIColorCode = map[Level]ANSIColorCode{
	ErrorLevel: RedCode,
	WarnLevel:  YellowCode,
	InfoLevel:  BlueCode,
	DoneLevel:  GreenCode,
	DebugLevel: MagentaCode,
}

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
	switch fields.Level {
	case ErrorLevel:
		l.print(fields.Level, message)
	case WarnLevel:
		l.print(fields.Level, message)
	case InfoLevel:
		l.print(fields.Level, message)
	case DoneLevel:
		l.print(fields.Level, message)
	case DebugLevel:
		l.print(fields.Level, message)
	default:
		l.print(fields.Level, message)
	}
}

func (l *consoleLogger) print(level Level, message string) {
	message = createLogMsg(level, message)
	if _, err := fmt.Fprint(l.output, message); err != nil {
		// Encountered an error during writing the message to the output. Manually construct a message for
		// the error and print it to the stdout.
		fmt.Printf("writing log message failed: %s", err)
	}
}

func createLogMsg(level Level, message string) string {
	color := levelToANSIColorCode[level]
	if color != "" {
		return addColor(color, message)
	}
	return message
}

func addColor(color ANSIColorCode, msg string) string {
	return string(color) + msg + string(ResetCode)
}
