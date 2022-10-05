package logger

import (
	"io"
	"time"
)

var globalLogger Logger

// InitGlobalLogger ...
func InitGlobalLogger(t LoggerType, producer Producer, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) {
	globalLogger = NewLogger(t, producer, out, debugLogEnabled, timeProvider)
}

// Error ...
func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Warn ...
func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Info ...
func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

// Infof ...
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Done ...
func Done(args ...interface{}) {
	globalLogger.Done(args...)
}

// Donef ...
func Donef(format string, args ...interface{}) {
	globalLogger.Donef(format, args...)
}

// Print ...
func Print(args ...interface{}) {
	globalLogger.Print(args...)
}

// Printf ...
func Printf(format string, args ...interface{}) {
	globalLogger.Printf(format, args...)
}

// Debug ...
func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}
