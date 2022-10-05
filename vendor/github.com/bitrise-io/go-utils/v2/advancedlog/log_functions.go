package logger

import (
	"io"
	"os"
	"time"
)

var globalLogger Logger

func getGlobalLogger() Logger {
	if globalLogger == nil {
		globalLogger = NewLogger(ConsoleLogger, BitriseCLI, os.Stdout, false, time.Now)
	}
	return globalLogger
}

// InitGlobalLogger ...
func InitGlobalLogger(t LoggerType, producer Producer, out io.Writer, debugLogEnabled bool, timeProvider func() time.Time) {
	globalLogger = NewLogger(t, producer, out, debugLogEnabled, timeProvider)
}

// Error ...
func Error(args ...interface{}) {
	getGlobalLogger().Error(args...)
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	getGlobalLogger().Errorf(format, args...)
}

// Warn ...
func Warn(args ...interface{}) {
	getGlobalLogger().Warn(args...)
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	getGlobalLogger().Warnf(format, args...)
}

// Info ...
func Info(args ...interface{}) {
	getGlobalLogger().Info(args...)
}

// Infof ...
func Infof(format string, args ...interface{}) {
	getGlobalLogger().Infof(format, args...)
}

// Done ...
func Done(args ...interface{}) {
	getGlobalLogger().Done(args...)
}

// Donef ...
func Donef(format string, args ...interface{}) {
	getGlobalLogger().Donef(format, args...)
}

// Print ...
func Print(args ...interface{}) {
	getGlobalLogger().Print(args...)
}

// Printf ...
func Printf(format string, args ...interface{}) {
	getGlobalLogger().Printf(format, args...)
}

// Debug ...
func Debug(args ...interface{}) {
	getGlobalLogger().Debug(args...)
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	getGlobalLogger().Debugf(format, args...)
}
