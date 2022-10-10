package corelog_test

import (
	"os"
	"time"

	"github.com/bitrise-io/bitrise/log/corelog"
)

func referenceTime() time.Time {
	return time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC)
}

func ExampleLogger() {
	var logger corelog.Logger

	logger = corelog.NewLogger(corelog.JSONLogger, os.Stdout, referenceTime)
	logger.LogMessage("Debug message", corelog.MessageFields{
		Level:    corelog.DebugLevel,
		Producer: corelog.BitriseCLI,
	})

	logger = corelog.NewLogger(corelog.ConsoleLogger, os.Stdout, referenceTime)
	logger.LogMessage("Info message", corelog.MessageFields{
		Level:    corelog.InfoLevel,
		Producer: corelog.Step,
	})

	// Output: {"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"bitrise_cli","level":"debug","message":"Debug message"}
	// [34;1mInfo message[0m
}
