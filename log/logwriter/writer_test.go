package logwriter_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func referenceTime() time.Time {
	return time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC)
}

func Test_GivenWriter_WhenStdoutIsUsed_ThenCapturesTheOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		producer        log.Producer
		loggerType      log.LoggerType
		message         string
		expectedMessage string
	}{
		{
			name:            "ClI console log",
			producer:        log.BitriseCLI,
			loggerType:      log.ConsoleLogger,
			message:         "Test message",
			expectedMessage: "Test message",
		},
		{
			name:            "Step JSON log",
			producer:        log.Step,
			loggerType:      log.JSONLogger,
			message:         "Test message",
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"step","level":"normal","message":"Test message"}` + "\n",
		},
		{
			name:            "Empty step JSON log",
			producer:        log.Step,
			loggerType:      log.JSONLogger,
			message:         "",
			expectedMessage: "",
		},
		{
			name:            "New line step JSON log",
			producer:        log.Step,
			loggerType:      log.JSONLogger,
			message:         "\n",
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"step","level":"normal","message":"\n"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			opts := log.LoggerOpts{
				LoggerType:        tt.loggerType,
				Producer:          tt.producer,
				ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
				DebugLogEnabled:   true,
				Writer:            &buf,
				TimeProvider:      referenceTime,
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			b := []byte(tt.message)

			_, err := writer.Write(b)
			assert.NoError(t, err)
			require.Equal(t, tt.expectedMessage, buf.String())
		})
	}
}

func ExampleNewLogWriter() {
	opts := log.LoggerOpts{
		LoggerType:        log.JSONLogger,
		Producer:          log.BitriseCLI,
		ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
		DebugLogEnabled:   true,
		Writer:            os.Stdout,
		TimeProvider:      referenceTime,
	}
	logger := log.NewLogger(opts)
	writer := logwriter.NewLogWriter(logger)
	cmd := exec.Command("echo", "test")
	cmd.Stdout = writer
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	// Output: {"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"bitrise_cli","level":"normal","message":"test\n"}
}
