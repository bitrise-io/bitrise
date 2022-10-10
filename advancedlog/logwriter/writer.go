package logwriter

import (
	"io"
	"time"

	"github.com/bitrise-io/bitrise/advancedlog/corelog"
)

type LoggerType corelog.LoggerType

const (
	JSONLogger    = LoggerType(corelog.JSONLogger)
	ConsoleLogger = LoggerType(corelog.ConsoleLogger)
)

type Producer corelog.Producer

const (
	BitriseCLI = Producer(corelog.BitriseCLI)
	Step       = Producer(corelog.Step)
)

type LogWriterOpts struct {
	Producer   Producer
	ProducerID string
}

// LogWriter ...
type LogWriter struct {
	logger          corelog.Logger
	opts            LogWriterOpts
	debugLogEnabled bool
}

// NewLogWriter ...
func NewLogWriter(t LoggerType, opts LogWriterOpts, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) LogWriter {
	logger := corelog.NewLogger(corelog.LoggerType(t), out, timeProvider)
	return LogWriter{
		logger:          logger,
		opts:            opts,
		debugLogEnabled: debugLogEnabled,
	}
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	level, message := convertColoredString(string(p))
	if !w.debugLogEnabled && level == corelog.DebugLevel {
		return len(p), nil
	}

	w.logger.LogMessage(message, corelog.MessageFields{
		Level:      level,
		Producer:   corelog.Producer(w.opts.Producer),
		ProducerID: w.opts.ProducerID,
	})
	return len(p), nil
}
