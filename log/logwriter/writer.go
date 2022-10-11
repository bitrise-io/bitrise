package logwriter

import (
	"io"
	"time"

	"github.com/bitrise-io/bitrise/log/corelog"
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
	t               LoggerType
	logger          corelog.Logger
	opts            LogWriterOpts
	debugLogEnabled bool
	timeProvider    func() time.Time
}

// NewLogWriter ...
func NewLogWriter(t LoggerType, opts LogWriterOpts, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) LogWriter {
	logger := corelog.NewLogger(corelog.LoggerType(t), out)
	return LogWriter{
		t:               t,
		logger:          logger,
		opts:            opts,
		debugLogEnabled: debugLogEnabled,
		timeProvider:    timeProvider,
	}
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	level, message := convertColoredString(string(p))
	if !w.debugLogEnabled && level == corelog.DebugLevel {
		return len(p), nil
	}

	var fields corelog.MessageFields
	if w.t == JSONLogger {
		fields = corelog.CreateJSONLogMessageFields(corelog.Producer(w.opts.Producer), w.opts.ProducerID, level, w.timeProvider)
	} else {
		fields = corelog.CreateConsoleLogMessageFields(level, nil)
	}

	w.logger.LogMessage(message, fields)
	return len(p), nil
}
