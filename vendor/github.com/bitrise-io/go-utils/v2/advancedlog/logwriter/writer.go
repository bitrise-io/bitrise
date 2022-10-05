package logwriter

import (
	"io"
	"time"

	"github.com/bitrise-io/go-utils/v2/advancedlog/corelog"
)

type LoggerType corelog.LoggerType

const (
	JSONLogger    LoggerType = LoggerType(corelog.JSONLogger)
	ConsoleLogger LoggerType = LoggerType(corelog.ConsoleLogger)
)

type Producer corelog.Producer

const (
	BitriseCLI Producer = Producer(corelog.BitriseCLI)
	Step       Producer = Producer(corelog.Step)
)

// LogWriter ...
type LogWriter struct {
	logger          corelog.Logger
	producer        corelog.Producer
	debugLogEnabled bool
}

// NewLogWriter ...
func NewLogWriter(t LoggerType, producer Producer, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) LogWriter {
	logger := corelog.NewLogger(corelog.LoggerType(t), out, timeProvider)
	return LogWriter{
		logger:          logger,
		producer:        corelog.Producer(producer),
		debugLogEnabled: debugLogEnabled,
	}
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	level, message := convertColoredString(string(p))
	if !w.debugLogEnabled && level == corelog.DebugLevel {
		return len(p), nil
	}

	w.logger.LogMessage(w.producer, level, message)
	return len(p), nil
}
