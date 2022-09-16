package logger

import (
	"fmt"

	logutils "github.com/bitrise-io/go-utils/v2/log"
)

// MainLogger ...
type MainLogger struct {
	internalLogger  SimplifiedLogger
	debugLogEnabled bool
}

// NewMainLogger ...
func NewMainLogger() MainLogger {
	return MainLogger{
		internalLogger:  newLegacyLogger(logutils.NewLogger()),
		debugLogEnabled: false,
	}
}

// Debug ...
func (m *MainLogger) Debug(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, DebugLevel, fmt.Sprint(args...))
}

// Debugf ...
func (m *MainLogger) Debugf(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, DebugLevel, fmt.Sprintf(format, args...))
}

// Info ...
func (m *MainLogger) Info(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, InfoLevel, fmt.Sprint(args...))
}

// Infof ...
func (m *MainLogger) Infof(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, InfoLevel, fmt.Sprintf(format, args...))
}

// Done ...
func (m *MainLogger) Done(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, DoneLevel, fmt.Sprint(args...))
}

// Donef ...
func (m *MainLogger) Donef(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, DoneLevel, fmt.Sprintf(format, args...))
}

// Warn ...
func (m *MainLogger) Warn(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, WarnLevel, fmt.Sprint(args...))
}

// Warnf ...
func (m *MainLogger) Warnf(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, WarnLevel, fmt.Sprintf(format, args...))
}

// Error ...
func (m *MainLogger) Error(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, ErrorLevel, fmt.Sprint(args...))
}

// Errorf ...
func (m *MainLogger) Errorf(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatal ...
func (m *MainLogger) Fatal(args ...interface{}) {
	m.Error(args...)
}

// Fatalf ...
func (m *MainLogger) Fatalf(format string, args ...interface{}) {
	m.Errorf(format, args...)
}

// Print ...
func (m *MainLogger) Print(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, NormalLevel, fmt.Sprintln(args...))
}

// Printf ...
func (m *MainLogger) Printf(format string, args ...interface{}) {
	m.internalLogger.LogMessage(CLI, NormalLevel, fmt.Sprintf(format, args...))
}

// Println ...
func (m *MainLogger) Println(args ...interface{}) {
	m.internalLogger.LogMessage(CLI, NormalLevel, fmt.Sprintln(args...))
}

// EnableDebugLog ...
func (m *MainLogger) EnableDebugLog(enable bool) {
	m.internalLogger.EnableDebugLog(enable)
	m.debugLogEnabled = enable
}

// IsDebugLogEnabled ...
func (m *MainLogger) IsDebugLogEnabled() bool {
	return m.debugLogEnabled
}

func (m *MainLogger) setInternalLogger(logger SimplifiedLogger) {
	m.internalLogger = logger
	m.internalLogger.EnableDebugLog(m.debugLogEnabled)
}

// LogMessage ...
func (m *MainLogger) LogMessage(producer Producer, level Level, message string) {
	m.internalLogger.LogMessage(producer, level, message)
}
