package stepoutput

import (
	"bytes"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
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

			w := NewWriter(nil, opts)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			_, err := w.Flush()
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
			w := NewWriter(nil, opts)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			_, err := w.Flush()
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
			w := NewWriter([]string{"secret value"}, opts)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}

			_, err := w.Flush()
			require.NoError(t, err)

			require.Equal(t, tt.want, buff.String())
		})
	}
}
