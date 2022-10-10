package logger

import (
	"fmt"
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

// defaultLogger ...
type defaultLogger struct {
	logger          corelog.Logger
	opts            LoggerOpts
	debugLogEnabled bool
}

type LoggerOpts struct {
	Producer   Producer
	ProducerID string
}

// NewLogger ...
func NewLogger(t LoggerType, opts LoggerOpts, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) Logger {
	coreLogger := corelog.NewLogger(corelog.LoggerType(t), out, timeProvider)
	return &defaultLogger{
		logger:          coreLogger,
		opts:            opts,
		debugLogEnabled: debugLogEnabled,
	}
}

func (m *defaultLogger) logMessage(message string, level corelog.Level, opts LoggerOpts) {
	m.logger.LogMessage(message, corelog.MessageFields{
		Level:      level,
		Producer:   corelog.Producer(opts.Producer),
		ProducerID: opts.ProducerID,
	})
}

// Error ...
func (m *defaultLogger) Error(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.ErrorLevel, m.opts)
}

// Errorf ...
func (m *defaultLogger) Errorf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.ErrorLevel, m.opts)
}

// Warn ...
func (m *defaultLogger) Warn(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.WarnLevel, m.opts)
}

// Warnf ...
func (m *defaultLogger) Warnf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.WarnLevel, m.opts)
}

// Info ...
func (m *defaultLogger) Info(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.InfoLevel, m.opts)
}

// Infof ...
func (m *defaultLogger) Infof(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.InfoLevel, m.opts)
}

// Done ...
func (m *defaultLogger) Done(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DoneLevel, m.opts)
}

// Donef ...
func (m *defaultLogger) Donef(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DoneLevel, m.opts)
}

// Print ...
func (m *defaultLogger) Print(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.NormalLevel, m.opts)
}

// Printf ...
func (m *defaultLogger) Printf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.NormalLevel, m.opts)
}

// Debug ...
func (m *defaultLogger) Debug(args ...interface{}) {
	if !m.debugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DebugLevel, m.opts)
}

// Debugf ...
func (m *defaultLogger) Debugf(format string, args ...interface{}) {
	if !m.debugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DebugLevel, m.opts)
}
