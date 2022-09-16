package logger

import "fmt"

// Debug ...
func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Info ...
func Info(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, InfoLevel, fmt.Sprint(args...))
}

// Infof ...
func Infof(format string, args ...interface{}) {
	DefaultLogger.LogMessage(CLI, InfoLevel, fmt.Sprintf(format, args...))
}

// Done ...
func Done(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, DoneLevel, fmt.Sprint(args...))
}

// Donef ...
func Donef(format string, args ...interface{}) {
	DefaultLogger.LogMessage(CLI, DoneLevel, fmt.Sprintf(format, args...))
}

// Warn ...
func Warn(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, WarnLevel, fmt.Sprint(args...))
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	DefaultLogger.LogMessage(CLI, WarnLevel, fmt.Sprintf(format, args...))
}

// Error ...
func Error(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, ErrorLevel, fmt.Sprint(args...))
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	DefaultLogger.LogMessage(CLI, ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatal ...
func Fatal(args ...interface{}) {
	Error(args...)
}

// Fatalf ...
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
}

// Print ...
func Print(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, NormalLevel, fmt.Sprint(args...))
}

// Printf ...
func Printf(format string, args ...interface{}) {
	DefaultLogger.LogMessage(CLI, NormalLevel, fmt.Sprintf(format, args...))
}

// Println ...
func Println(args ...interface{}) {
	DefaultLogger.LogMessage(CLI, NormalLevel, fmt.Sprint(args...))
}

// IsDebugLogEnabled ...
func IsDebugLogEnabled() bool {
	return DefaultLogger.IsDebugLogEnabled()
}
