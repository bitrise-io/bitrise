package corelog

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GivenConsoleLogger_WhenLogMessageInvoked_ThenLogsItCorrectly(t *testing.T) {
	tests := []struct {
		name            string
		messageFields   MessageFields
		message         string
		expectedMessage string
	}{
		{
			name: "Info message with fields",
			messageFields: MessageFields{
				Timestamp:  "2022.01.01",
				Producer:   "step",
				ProducerID: "step--unique-id",
				Level:      InfoLevel,
			},
			message:         "Info message",
			expectedMessage: "[2022.01.01] step step--unique-id \u001B[34;1mInfo message\u001B[0m",
		},
		{
			name: "Empty message with fields",
			messageFields: MessageFields{
				Timestamp:  "2022.01.01",
				Producer:   "step",
				ProducerID: "step--unique-id",
				Level:      InfoLevel,
			},
			message:         "",
			expectedMessage: "[2022.01.01] step step--unique-id",
		},
		{
			name: "Error log",
			messageFields: MessageFields{
				Level: ErrorLevel,
			},
			message:         "Error",
			expectedMessage: "\u001B[31;1mError\u001B[0m",
		},
		{
			name: "Warning log",
			messageFields: MessageFields{
				Level: WarnLevel,
			},
			message:         "Warning",
			expectedMessage: "\u001B[33;1mWarning\u001B[0m",
		},
		{
			name: "Info log",
			messageFields: MessageFields{
				Level: InfoLevel,
			},
			message:         "Info",
			expectedMessage: "\u001B[34;1mInfo\u001B[0m",
		},
		{
			name: "Done log",
			messageFields: MessageFields{
				Level: DoneLevel,
			},
			message:         "Done",
			expectedMessage: "\u001B[32;1mDone\u001B[0m",
		},
		{
			name: "Normal log",
			messageFields: MessageFields{
				Level: NormalLevel,
			},
			message:         "Normal",
			expectedMessage: "Normal",
		},
		{
			name: "Debug log",
			messageFields: MessageFields{
				Level: DebugLevel,
			},
			message:         "Debug",
			expectedMessage: "\u001B[35;1mDebug\u001B[0m",
		},
		{
			name: "Empty message is logged",
			messageFields: MessageFields{
				Level: InfoLevel,
			},
			message:         "",
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer

			logger := newConsoleLogger(&buff)
			logger.LogMessage(tt.message, tt.messageFields)

			require.Equal(t, tt.expectedMessage, buff.String())
		})
	}
}
