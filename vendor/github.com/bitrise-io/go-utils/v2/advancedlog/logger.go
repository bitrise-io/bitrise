package logger

import "os"

// OutputFormat ...
type OutputFormat string

const (
	// JSONFormat ...
	JSONFormat OutputFormat = "json"
)

// Logger ...
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Done(args ...interface{})
	Donef(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	EnableDebugLog(enable bool)
	IsDebugLogEnabled() bool
}

// SimplifiedLogger ...
type SimplifiedLogger interface {
	EnableDebugLog(enabled bool)
	IsDebugLogEnabled() bool
	LogMessage(producer Producer, level Level, message string)
}

// DefaultLogger ...
var DefaultLogger = NewMainLogger()

// SetOutputFormat ...
func SetOutputFormat(outputFormat string) {
	if OutputFormat(outputFormat) == JSONFormat {
		DefaultLogger.setInternalLogger(newJSONLogger(os.Stdout, defaultTimeProvider))
	}
}

// SetEnableDebugLog ...
func SetEnableDebugLog(enable bool) {
	DefaultLogger.EnableDebugLog(enable)
}
