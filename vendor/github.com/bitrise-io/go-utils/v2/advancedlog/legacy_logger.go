package logger

import (
	logutils "github.com/bitrise-io/go-utils/v2/log"
)

type legacyLogger struct {
	debugLogEnabled bool
	logger          logutils.Logger
}

func newLegacyLogger(logger logutils.Logger) SimplifiedLogger {
	return &legacyLogger{
		debugLogEnabled: false,
		logger:          logger,
	}
}

// EnableDebugLog ...
func (l *legacyLogger) EnableDebugLog(enabled bool) {
	l.logger.EnableDebugLog(enabled)
	l.debugLogEnabled = enabled
}

// IsDebugLogEnabled ...
func (l *legacyLogger) IsDebugLogEnabled() bool {
	return l.debugLogEnabled
}

// LogMessage ...
func (l *legacyLogger) LogMessage(producer Producer, level Level, message string) {
	if !l.debugLogEnabled && level == DebugLevel {
		return
	}

	// All the print functions below automatically add a newline to the end. We need to replace the newline with an
	// empty line not to log multiple newlines.
	if message == "\n" {
		message = ""
	}

	switch level {
	case ErrorLevel:
		l.logger.Errorf(message)
	case WarnLevel:
		l.logger.Warnf(message)
	case InfoLevel:
		l.logger.Infof(message)
	case DoneLevel:
		l.logger.Donef(message)
	case DebugLevel:
		l.logger.Debugf(message)
	default:
		l.logger.Printf(message)
	}
}
