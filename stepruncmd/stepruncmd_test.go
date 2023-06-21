package stepruncmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/bitrise/stepruncmd/filterwriter"
	"github.com/stretchr/testify/require"
)

func Test_GivenWriter_WhenConsoleLogging_ThenTransmitsLogs(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     string
	}{
		{
			name:     "Simple log",
			messages: []string{"failed to create file artifact: /bitrise/src/assets"},
			want:     "failed to create file artifact: /bitrise/src/assets",
		},
		{
			name: "Error log",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			want: `[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			opts := log.LoggerOpts{
				LoggerType: log.ConsoleLogger,
				Writer:     &buff,
				TimeProvider: func() time.Time {
					return time.Time{}
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			w := NewStdoutWriter(nil, writer)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			err := w.Close()
			require.NoError(t, err)

			require.Equal(t, tt.want, buff.String())
		})
	}
}

func Test_GivenWriter_WhenJSONLogging_ThenWritesJSON(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     string
	}{
		{
			name:     "Simple log",
			messages: []string{"failed to create file artifact: /bitrise/src/assets"},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"normal","message":"failed to create file artifact: /bitrise/src/assets"}
`,
		},
		{
			name: "Error log",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets"}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			opts := log.LoggerOpts{
				LoggerType: log.JSONLogger,
				Producer:   "Test",
				ProducerID: "UUID",
				Writer:     &buff,
				TimeProvider: func() time.Time {
					return time.Time{}
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			w := NewStdoutWriter(nil, writer)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			err := w.Close()
			require.NoError(t, err)

			require.Equal(t, tt.want, buff.String())
		})
	}
}

func Test_GivenWriter_WhenJSONLoggingAndSecretFiltering_ThenWritesJSON(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     string
	}{
		{
			name:     "Simple log",
			messages: []string{"failed to create file artifact: /bitrise/src/assets\n"},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"normal","message":"failed to create file artifact: /bitrise/src/assets\n"}
`,
		},
		{
			name: "Error log",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m
`},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets\n"}
`,
		},
		{
			name: "Error log in multiple messages",
			messages: []string{"[31;1mfailed to create file artifact: /bitrise/src/assets:\n",
				"  failed to get file size, error: file not exist at: /bitrise/src/assets[0m"},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets"}
`,
		},
		{
			name: "Error log in multiple messages, color reset after newline character",
			messages: []string{"[31;1mfailed to create file artifact: /bitrise/src/assets:\n",
				"  failed to get file size, error: file not exist at: /bitrise/src/assets\n",
				"[0m"},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"Test","producer_id":"UUID","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets\n"}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			opts := log.LoggerOpts{
				LoggerType: log.JSONLogger,
				Producer:   "Test",
				ProducerID: "UUID",
				Writer:     &buff,
				TimeProvider: func() time.Time {
					return time.Time{}
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			w := NewStdoutWriter([]string{"secret value"}, writer)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			err := w.Close()
			require.NoError(t, err)

			require.Equal(t, tt.want, buff.String())
		})
	}
}

func Test_GivenWriter_WhenJSONLoggingAndSecretFiltering_ThenReturnsError(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     []string
	}{
		{
			name:     "Simple log",
			messages: []string{"failed to create file artifact: /bitrise/src/assets"},
			want:     nil,
		},
		{
			name: "Error log",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			want: []string{"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			opts := log.LoggerOpts{
				LoggerType: log.JSONLogger,
				Producer:   "Test",
				ProducerID: "UUID",
				Writer:     &buff,
				TimeProvider: func() time.Time {
					// UnixNano() is 0 for this time
					return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
				},
			}
			logger := log.NewLogger(opts)
			writer := logwriter.NewLogWriter(logger)

			w := NewStdoutWriter([]string{"secret value"}, writer)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			err := w.Close()
			require.NoError(t, err)

			errors := w.ErrorMessages()
			require.Equal(t, tt.want, errors)
		})
	}
}

func Test_WhenSecretsProvided_ThenRootWriterIsFilterWriter(t *testing.T) {
	w := NewStdoutWriter([]string{"secret"}, nil)
	_, isFilterWriter := w.writer.(*filterwriter.Writer)
	require.True(t, isFilterWriter)
}
