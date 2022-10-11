package log

import (
	"fmt"
	"io"
	"time"

	"github.com/bitrise-io/bitrise/log/corelog"
)

// RFC3339MicroTimeLayout ...
const RFC3339MicroTimeLayout = "2006-01-02T15:04:05.999999Z07:00"

// ConsoleTimeLayout ...
const ConsoleTimeLayout = "15:04:05"

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
	opts   LoggerOpts
	logger corelog.Logger
}

type ConsoleLoggerOpts struct {
	Timestamp bool
}

type LoggerOpts struct {
	LoggerType        LoggerType
	Producer          Producer
	ProducerID        string
	ConsoleLoggerOpts ConsoleLoggerOpts
	DebugLogEnabled   bool
	Writer            io.Writer
	TimeProvider      func() time.Time
}

// NewLogger ...
func NewLogger(opts LoggerOpts) Logger {
	return newLogger(opts)
}

func newLogger(opts LoggerOpts) *defaultLogger {
	logger := corelog.NewLogger(corelog.LoggerType(opts.LoggerType), opts.Writer)
	return &defaultLogger{
		opts:   opts,
		logger: logger,
	}
}

// Error ...
func (m *defaultLogger) Error(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.ErrorLevel)
}

// Errorf ...
func (m *defaultLogger) Errorf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.ErrorLevel)
}

// Warn ...
func (m *defaultLogger) Warn(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.WarnLevel)
}

// Warnf ...
func (m *defaultLogger) Warnf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.WarnLevel)
}

// Info ...
func (m *defaultLogger) Info(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.InfoLevel)
}

// Infof ...
func (m *defaultLogger) Infof(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.InfoLevel)
}

// Done ...
func (m *defaultLogger) Done(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DoneLevel)
}

// Donef ...
func (m *defaultLogger) Donef(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DoneLevel)
}

// Print ...
func (m *defaultLogger) Print(args ...interface{}) {
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.NormalLevel)
}

// Printf ...
func (m *defaultLogger) Printf(format string, args ...interface{}) {
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.NormalLevel)
}

// Debug ...
func (m *defaultLogger) Debug(args ...interface{}) {
	if !m.opts.DebugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprint(args...)+"\n", corelog.DebugLevel)
}

// Debugf ...
func (m *defaultLogger) Debugf(format string, args ...interface{}) {
	if !m.opts.DebugLogEnabled {
		return
	}
	m.logMessage(fmt.Sprintf(format, args...)+"\n", corelog.DebugLevel)
}

// LogMessage ...
func (m *defaultLogger) LogMessage(message string, level corelog.Level) {
	if level == corelog.DebugLevel && !m.opts.DebugLogEnabled {
		return
	}

	m.logMessage(message, level)
}

func (m *defaultLogger) logMessage(message string, level corelog.Level) {
	fields := m.createMessageFields(level)
	m.logger.LogMessage(message, corelog.MessageFields(fields))
}

func (m *defaultLogger) createMessageFields(level corelog.Level) MessageFields {
	if m.opts.LoggerType == JSONLogger {
		return createJSONLogMessageFields(m.opts.Producer, m.opts.ProducerID, level, m.opts.TimeProvider)
	}

	var tProvider func() time.Time
	if m.opts.ConsoleLoggerOpts.Timestamp {
		tProvider = m.opts.TimeProvider
	}
	return createConsoleLogMessageFields(level, tProvider)
}

func createJSONLogMessageFields(producer Producer, producerID string, level corelog.Level, timeProvider func() time.Time) MessageFields {
	return MessageFields{
		Timestamp:  timeProvider().Format(RFC3339MicroTimeLayout),
		Producer:   corelog.Producer(producer),
		ProducerID: producerID,
		Level:      level,
	}
}

func createConsoleLogMessageFields(level corelog.Level, timeProvider func() time.Time) MessageFields {
	fields := MessageFields{
		Level: level,
	}
	if timeProvider != nil {
		fields.Timestamp = timeProvider().Format(ConsoleTimeLayout)
	}
	return fields
}
