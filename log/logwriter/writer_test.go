package logwriter_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/corelog"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func referenceTime() time.Time {
	return time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC)
}

func Test_GivenWriter_WhenStdoutIsUsed_ThenCapturesTheOutput(t *testing.T) {
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
			writer := logwriter.NewLogLevelWriter(logger)

			b := []byte(tt.message)

			_, err := writer.Write(b)
			assert.NoError(t, err)
			require.Equal(t, tt.expectedMessage, buf.String())
		})
	}
}

func Test_GivenWriter_WhenMessageIsWritten_ThenParsesLogLevel(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedLevel   corelog.Level
		expectedMessage string
	}{
		{
			name:          "Normal message without a color literal",
			message:       "This is a normal message without a color literal\n",
			expectedLevel: corelog.NormalLevel,
			//expectedMessage: "This is a normal message without a color literal\n",
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"normal","message":"This is a normal message without a color literal\n"}` + "\n",
		},

		{
			name:            "Error message",
			message:         "\u001B[31;1mThis is an error\u001B[0m",
			expectedLevel:   corelog.ErrorLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"error","message":"This is an error"}` + "\n",
		},
		{
			name:            "Warn message",
			message:         "\u001B[33;1mThis is a warning\u001B[0m",
			expectedLevel:   corelog.WarnLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"warn","message":"This is a warning"}` + "\n",
		},
		{
			name:            "Info message",
			message:         "\u001B[34;1mThis is an Info\u001B[0m",
			expectedLevel:   corelog.InfoLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"info","message":"This is an Info"}` + "\n",
		},
		{
			name:            "Done message",
			message:         "\u001B[32;1mThis is a done message\u001B[0m",
			expectedLevel:   corelog.DoneLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"done","message":"This is a done message"}` + "\n",
		},
		{
			name:            "Debug message",
			message:         "\u001B[35;1mThis is a debug message\u001B[0m",
			expectedLevel:   corelog.DebugLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"debug","message":"This is a debug message"}` + "\n",
		},
		{
			name:            "Error message with whitespaces at the end",
			message:         "\u001B[31;1mLast error\u001B[0m   \n",
			expectedLevel:   corelog.ErrorLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"error","message":"Last error   \n"}` + "\n",
		},
		{
			name:            "Error message with whitespaces at the beginning",
			message:         "  \u001B[31;1mLast error\u001B[0m   \n",
			expectedLevel:   corelog.NormalLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"normal","message":"  \u001b[31;1mLast error\u001b[0m   \n"}` + "\n",
		},
		{
			name:            "Error message without a closing color literal",
			message:         "\u001B[31;1mAnother error\n",
			expectedLevel:   corelog.NormalLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"normal","message":"\u001b[31;1mAnother error\n\u001b[31;1mAnother error\n"}` + "\n",
		},
		{
			name:            "Info message with multiple embedded colors",
			message:         "\u001B[34;1mThis is \u001B[33;1mmulti color \u001B[31;1mInfo message\u001B[0m",
			expectedLevel:   corelog.NormalLevel,
			expectedMessage: `{"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"","level":"normal","message":"\u001b[34;1mThis is \u001b[33;1mmulti color \u001b[31;1mInfo message\u001b[0m"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			opts := log.LoggerOpts{
				LoggerType:        log.JSONLogger,
				ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
				DebugLogEnabled:   true,
				Writer:            &buf,
				TimeProvider:      referenceTime,
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogLevelWriter(logger)

			b := []byte(tt.message)

			_, err := writer.Write(b)
			assert.NoError(t, err)
			_, err = writer.Flush()
			assert.NoError(t, err)
			require.Equal(t, tt.expectedMessage, buf.String())
		})
	}
}

func ExampleNewLogLevelWriter() {
	opts := log.LoggerOpts{
		LoggerType:        log.JSONLogger,
		Producer:          log.BitriseCLI,
		ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
		DebugLogEnabled:   true,
		Writer:            os.Stdout,
		TimeProvider:      referenceTime,
	}
	logger := log.NewLogger(opts)
	writer := logwriter.NewLogLevelWriter(logger)
	cmd := exec.Command("echo", "test")
	cmd.Stdout = writer
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	// Output: {"timestamp":"2022-01-01T01:01:01Z","type":"log","producer":"bitrise_cli","level":"normal","message":"test\n"}
}
