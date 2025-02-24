package log_test

import (
	"os"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
)

func referenceTime() time.Time {
	return time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC)
}

func ExampleLogger() {
	var logger log.Logger

	logger = log.NewLogger(log.LoggerOpts{
		LoggerType:      log.ConsoleLogger,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: true,
		Writer:          os.Stdout,
		TimeProvider:    referenceTime,
	})
	logger.Errorf("This is an %s", "error")

	logger = log.NewLogger(log.LoggerOpts{
		LoggerType:      log.JSONLogger,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: true,
		Writer:          os.Stdout,
		TimeProvider:    referenceTime,
	})
	logger.Debug("This is a debug message")

	log.InitGlobalLogger(log.LoggerOpts{
		LoggerType:      log.JSONLogger,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: true,
		Writer:          os.Stdout,
		TimeProvider:    referenceTime,
	})
	log.Info("This is an info message")
}
