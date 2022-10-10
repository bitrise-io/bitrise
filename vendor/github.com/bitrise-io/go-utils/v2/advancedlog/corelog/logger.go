package corelog

import (
	"io"
	"time"
)

type LoggerType string

const (
	JSONLogger    LoggerType = "json"
	ConsoleLogger LoggerType = "console"
)

// MessageFields ...
type MessageFields struct {
	Level      Level
	Producer   Producer
	ProducerID string
}

// Logger ...
type Logger interface {
	LogMessage(message string, fields MessageFields)
}

func NewLogger(t LoggerType, output io.Writer, timeProvider func() time.Time) Logger {
	switch t {
	case JSONLogger:
		return newJSONLogger(output, timeProvider)
	default:
		return newConsoleLogger(output)
	}
}
