package log

import (
	"os"
	"time"

	"github.com/bitrise-io/bitrise/v2/log/corelog"
	"github.com/bitrise-io/bitrise/v2/models"
)

var globalLogger *defaultLogger

func getGlobalLogger() Logger {
	if globalLogger == nil {
		opts := LoggerOpts{
			LoggerType:   ConsoleLogger,
			Producer:     BitriseCLI,
			Writer:       os.Stdout,
			TimeProvider: time.Now,
		}
		globalLogger = newLogger(opts)
	}
	return globalLogger
}

// GetGlobalLoggerOpts ...
func GetGlobalLoggerOpts() LoggerOpts {
	getGlobalLogger()
	return globalLogger.opts
}

// InitGlobalLogger ...
func InitGlobalLogger(opts LoggerOpts) {
	globalLogger = newLogger(opts)
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

// LogMessage ...
func LogMessage(message string, level corelog.Level) {
	getGlobalLogger().LogMessage(message, level)
}

func PrintBitriseStartedEvent(plan models.WorkflowRunPlan) {
	getGlobalLogger().PrintBitriseStartedEvent(plan)
}

func PrintStepStartedEvent(params StepStartedParams) {
	getGlobalLogger().PrintStepStartedEvent(params)
}

func PrintStepFinishedEvent(params StepFinishedParams) {
	getGlobalLogger().PrintStepFinishedEvent(params)
}
