package logwriter_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GivenWriter(t *testing.T) {
	tests := []struct {
		name            string
		loggerType      log.LoggerType
		message         string
		expectedMessage string
	}{
		{
			name:            "Empty message, console logging",
			loggerType:      log.ConsoleLogger,
			message:         "",
			expectedMessage: "",
		},
		{
			name:            "New line message, console logging",
			loggerType:      log.ConsoleLogger,
			message:         "\n",
			expectedMessage: "\n",
		},
		{
			name:            "Empty message, JSON logging",
			loggerType:      log.JSONLogger,
			message:         "",
			expectedMessage: "",
		},
		{
			name:            "New line message, json logging",
			loggerType:      log.JSONLogger,
			message:         "\n",
			expectedMessage: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"\n"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			opts := log.LoggerOpts{
				LoggerType:        tt.loggerType,
				ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
				DebugLogEnabled:   true,
				Writer:            &buf,
				TimeProvider: func() time.Time {
					return time.Time{}
				},
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

func Test_GivenWriter_WhenJSONLogging_ThenDetectsLogLevel(t *testing.T) {
	tests := []struct {
		name             string
		messages         []string
		expectedMessages []string
	}{
		{
			name:             "Writes messages with normal log level by default",
			messages:         []string{"Hello Bitrise!"},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"Hello Bitrise!"}` + "\n"},
		},
		{
			name:             "Detects log level in a message",
			messages:         []string{"\u001B[34;1mLogin to the service\u001B[0m"},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"info","message":"Login to the service"}` + "\n"},
		},
		{
			name: "Detects a log level in a message stream",
			messages: []string{
				"\u001B[35;1mdetected login method:",
				"- API key",
				"- username\u001B[0m",
			},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"debug","message":"detected login method:\n- API key\n- username"}` + "\n"},
		},
		{
			name: "Detects multiple messages with log level in the message stream",
			messages: []string{
				"Hello Bitrise!",
				"\u001B[35;1mdetected login method:",
				"- API key",
				"- username\u001B[0m",
				"\u001B[34;1mLogin to the service\u001B[0m",
			},
			expectedMessages: []string{
				`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"Hello Bitrise!"}` + "\n",
				`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"debug","message":"detected login method:\n- API key\n- username"}` + "\n",
				`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"info","message":"Login to the service"}` + "\n",
			},
		},
		{
			name:             "Error message with whitespaces at the end (not a message with log level)",
			messages:         []string{"\u001B[31;1mLast error\u001B[0m   \n"},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"\u001b[31;1mLast error\u001b[0m   \n"}` + "\n"},
		},
		{
			name:             "Error message with whitespaces at the beginning (not a message with log level)",
			messages:         []string{"  \u001B[31;1mLast error\u001B[0m   \n"},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"  \u001b[31;1mLast error\u001b[0m   \n"}` + "\n"},
		},
		{
			name:             "Error message without a closing color literal (not a message with log level)",
			messages:         []string{"\u001B[31;1mAnother error\n"},
			expectedMessages: []string{`{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"normal","message":"\u001b[31;1mAnother error\n"}` + "\n"},
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
				TimeProvider: func() time.Time {
					return time.Time{}
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			var actualMessages []string
			for _, message := range tt.messages {
				b := []byte(message)

				_, err := writer.Write(b)
				assert.NoError(t, err)

				actualMessage := buf.String()
				buf.Reset()
				if actualMessage != "" {
					actualMessages = append(actualMessages, actualMessage)
				}
			}

			err := writer.Close()
			require.NoError(t, err)

			actualMessage := buf.String()
			buf.Reset()
			if actualMessage != "" {
				actualMessages = append(actualMessages, actualMessage)
			}

			require.Equal(t, tt.expectedMessages, actualMessages)
		})
	}
}

func Test_GivenWriter_WhenConsoleLogging_ThenDetectsLogLevel(t *testing.T) {
	tests := []struct {
		name             string
		messages         []string
		expectedMessages []string
	}{
		{
			name:             "Writes messages without log level as it is",
			messages:         []string{"Hello Bitrise!"},
			expectedMessages: []string{"Hello Bitrise!"},
		},
		{
			name:             "Writes messages with log level as it is",
			messages:         []string{"\u001B[34;1mLogin to the service\u001B[0m"},
			expectedMessages: []string{"\u001B[34;1mLogin to the service\u001B[0m"},
		},
		{
			name: "Detects a message with log level in the message stream",
			messages: []string{
				"\u001B[35;1mdetected login method:",
				"- API key",
				"- username\u001B[0m",
			},
			expectedMessages: []string{"\u001B[35;1mdetected login method:\n- API key\n- username\u001B[0m"},
		},
		{
			name: "Detects multiple messages with log level in the message stream",
			messages: []string{
				"Hello Bitrise!",
				"\u001B[35;1mdetected login method:",
				"- API key",
				"- username\u001B[0m",
				"\u001B[34;1mLogin to the service\u001B[0m",
			},
			expectedMessages: []string{
				"Hello Bitrise!",
				"\u001B[35;1mdetected login method:\n- API key\n- username\u001B[0m",
				"\u001B[34;1mLogin to the service\u001B[0m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			opts := log.LoggerOpts{
				ConsoleLoggerOpts: log.ConsoleLoggerOpts{},
				DebugLogEnabled:   true,
				Writer:            &buf,
				TimeProvider: func() time.Time {
					return time.Time{}
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			var actualMessages []string
			for _, message := range tt.messages {
				b := []byte(message)

				_, err := writer.Write(b)
				assert.NoError(t, err)

				actualMessage := buf.String()
				buf.Reset()
				if actualMessage != "" {
					actualMessages = append(actualMessages, actualMessage)
				}
			}

			err := writer.Close()
			require.NoError(t, err)

			require.Equal(t, tt.expectedMessages, actualMessages)
		})
	}
}
