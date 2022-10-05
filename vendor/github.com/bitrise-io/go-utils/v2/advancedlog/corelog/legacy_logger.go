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

type legacyLogger struct {
	output io.Writer
}

func newLegacyLogger(output io.Writer) *legacyLogger {
	return &legacyLogger{
		output: output,
	}
}

// LogMessage ...
func (l *legacyLogger) LogMessage(producer Producer, level Level, message string) {
	switch level {
	case ErrorLevel:
		l.print(level, message)
	case WarnLevel:
		l.print(level, message)
	case InfoLevel:
		l.print(level, message)
	case DoneLevel:
		l.print(level, message)
	case DebugLevel:
		l.print(level, message)
	default:
		l.print(level, message)
	}
}

func (l *legacyLogger) print(level Level, message string) {
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
