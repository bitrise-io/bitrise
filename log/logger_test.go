package log_test

import (
	"os"
	"time"

	"github.com/bitrise-io/bitrise/log"
)

func referenceTime() time.Time {
	return time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC)
}

func ExampleLogger() {
	var logger log.Logger

	logger = log.NewLogger(log.ConsoleLogger, log.LoggerOpts{Producer: log.BitriseCLI}, os.Stdout, true, referenceTime)
	logger.Errorf("This is an %s", "error")

	logger = log.NewLogger(log.JSONLogger, log.LoggerOpts{Producer: log.BitriseCLI}, os.Stdout, true, referenceTime)
	logger.Debug("This is a debug message")

	log.InitGlobalLogger(log.JSONLogger, log.LoggerOpts{Producer: log.BitriseCLI}, os.Stdout, true, referenceTime)
	log.Info("This is an info message")
}
