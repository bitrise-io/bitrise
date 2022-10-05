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
	producer        corelog.Producer
	debugLogEnabled bool
}

// NewLogger ...
func NewLogger(t LoggerType, producer Producer, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) Logger {
	coreLogger := corelog.NewLogger(corelog.LoggerType(t), out, timeProvider)
	return &defaultLogger{
		logger:          coreLogger,
		producer:        corelog.Producer(producer),
		debugLogEnabled: debugLogEnabled,
	}
}

// Error ...
func (m *defaultLogger) Error(args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.ErrorLevel, fmt.Sprint(args...)+"\n")
}

// Errorf ...
func (m *defaultLogger) Errorf(format string, args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.ErrorLevel, fmt.Sprintf(format, args...)+"\n")
}

// Warn ...
func (m *defaultLogger) Warn(args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.WarnLevel, fmt.Sprint(args...)+"\n")
}

// Warnf ...
func (m *defaultLogger) Warnf(format string, args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.WarnLevel, fmt.Sprintf(format, args...)+"\n")
}

// Info ...
func (m *defaultLogger) Info(args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.InfoLevel, fmt.Sprint(args...)+"\n")
}

// Infof ...
func (m *defaultLogger) Infof(format string, args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.InfoLevel, fmt.Sprintf(format, args...)+"\n")
}

// Done ...
func (m *defaultLogger) Done(args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.DoneLevel, fmt.Sprint(args...)+"\n")
}

// Donef ...
func (m *defaultLogger) Donef(format string, args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.DoneLevel, fmt.Sprintf(format, args...)+"\n")
}

// Print ...
func (m *defaultLogger) Print(args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.NormalLevel, fmt.Sprint(args...)+"\n")
}

// Printf ...
func (m *defaultLogger) Printf(format string, args ...interface{}) {
	m.logger.LogMessage(m.producer, corelog.NormalLevel, fmt.Sprintf(format, args...)+"\n")
}

// Debug ...
func (m *defaultLogger) Debug(args ...interface{}) {
	if !m.debugLogEnabled {
		return
	}
	m.logger.LogMessage(m.producer, corelog.DebugLevel, fmt.Sprint(args...)+"\n")
}

// Debugf ...
func (m *defaultLogger) Debugf(format string, args ...interface{}) {
	if !m.debugLogEnabled {
		return
	}
	m.logger.LogMessage(m.producer, corelog.DebugLevel, fmt.Sprintf(format, args...)+"\n")
}
