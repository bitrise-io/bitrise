package corelog

import (
	"io"
)

type LoggerType string

const (
	JSONLogger    LoggerType = "json"
	ConsoleLogger LoggerType = "console"
)

// Producer ...
type Producer string

const (
	// BitriseCLI ...
	BitriseCLI Producer = "bitrise_cli"
	// Step ...
	Step Producer = "step"
)

// Level ...
type Level string

const (
	// ErrorLevel ...
	ErrorLevel Level = "error"
	// WarnLevel ...
	WarnLevel Level = "warn"
	// InfoLevel ...
	InfoLevel Level = "info"
	// DoneLevel ...
	DoneLevel Level = "done"
	// NormalLevel ...
	NormalLevel Level = "normal"
	// DebugLevel ...
	DebugLevel Level = "debug"
)

// MessageLogFields ...
type MessageLogFields struct {
	Timestamp  string   `json:"timestamp"`
	Producer   Producer `json:"producer"`
	ProducerID string   `json:"producer_id,omitempty"`
	Level      Level    `json:"level"`
}

// EventLogFields ...
type EventLogFields struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
}

// Logger ...
type Logger interface {
	LogMessage(message string, fields MessageLogFields)
	LogEvent(content interface{}, fields EventLogFields)
}

func NewLogger(t LoggerType, output io.Writer) Logger {
	switch t {
	case JSONLogger:
		return newJSONLogger(output)
	default:
		return newConsoleLogger(output)
	}
}
