package corelog_test

import (
	"os"

	"github.com/bitrise-io/bitrise/log/corelog"
)

func ExampleLogger() {
	var logger corelog.Logger

	fields := corelog.MessageFields{
		Timestamp: "2022-01-01T01:01:01Z",
		Producer:  corelog.BitriseCLI,
		Level:     corelog.InfoLevel,
	}
	message := "Info message"

	logger = corelog.NewLogger(corelog.JSONLogger, os.Stdout)
	logger.LogMessage(message, fields)

	logger = corelog.NewLogger(corelog.ConsoleLogger, os.Stdout)
	logger.LogMessage(message, fields)

	// Output: {"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"bitrise_cli","level":"info","message":"Info message"}
	// [2022-01-01T01:01:01Z] bitrise_cli [34;1mInfo message[0m
}
