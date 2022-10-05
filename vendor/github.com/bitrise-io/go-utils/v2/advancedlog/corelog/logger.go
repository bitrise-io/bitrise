package corelog

import (
	"io"
	"time"
)

// Logger ...
type Logger interface {
	LogMessage(producer Producer, level Level, message string)
}

type LoggerType string

const (
	JSONLogger    LoggerType = "json"
	ConsoleLogger LoggerType = "console"
)

func NewLogger(t LoggerType, output io.Writer, timeProvider func() time.Time) Logger {
	switch t {
	case JSONLogger:
		return newJSONLogger(output, timeProvider)
	default:
		return newLegacyLogger(output)
	}
}
