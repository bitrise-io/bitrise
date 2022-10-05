package corelog

import (
	"fmt"
	"io"
)

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
		l.printf(level, message)
	case WarnLevel:
		l.printf(level, message)
	case InfoLevel:
		l.printf(level, message)
	case DoneLevel:
		l.printf(level, message)
	case DebugLevel:
		l.printf(level, message)
	default:
		l.printf(level, message)
	}
}

func addColor(color ANSIColorCode, msg string) string {
	return string(color) + msg + string(ResetCode)
}

var levelToANSIColorCode = map[Level]ANSIColorCode{
	ErrorLevel: RedCode,
	WarnLevel:  YellowCode,
	InfoLevel:  BlueCode,
	DoneLevel:  GreenCode,
	DebugLevel: MagentaCode,
}

func (l *legacyLogger) createLogMsg(level Level, message string) string {
	color := levelToANSIColorCode[level]
	return addColor(color, message)
}

func (l *legacyLogger) printf(level Level, message string) {
	message = l.createLogMsg(level, message)
	if _, err := fmt.Fprintln(l.output, message); err != nil {
		// Encountered an error during writing the json message to the output. Manually construct a json message for
		// the error and print it to the output
		fmt.Printf("writing log message failed: %s", err)
	}
}
