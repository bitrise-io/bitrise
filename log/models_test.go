package log

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStepStartedEventSerialisesToTheExpectedJsonMessage(t *testing.T) {
	tests := []struct {
		name           string
		params         StepStartedParams
		expectedOutput string
	}{
		{
			name: "Only the required json output fields are serialised",
			params: StepStartedParams{
				ExecutionId: "ExecutionId",
				Position:    0,
				IdVersion:   "IdVersion",
				Id:          "Id",
				Version:     "Version",
				Title:       "Title",
				Collection:  "Collection",
				Toolkit:     "Toolkit",
				StartTime:   "This is not needed",
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"idx\":0,\"id_version\":\"IdVersion\",\"id\":\"Id\",\"version\":\"Version\",\"title\":\"Title\",\"collection\":\"Collection\",\"toolkit\":\"Toolkit\"}",
		},
		{
			name: "Empty fields are kept",
			params: StepStartedParams{
				ExecutionId: "ExecutionId",
				Position:    0,
				IdVersion:   "IdVersion",
				Id:          "Id",
				Version:     "",
				Title:       "",
				Collection:  "Collection",
				Toolkit:     "",
				StartTime:   "",
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"idx\":0,\"id_version\":\"IdVersion\",\"id\":\"Id\",\"version\":\"\",\"title\":\"\",\"collection\":\"Collection\",\"toolkit\":\"\"}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := json.Marshal(test.params)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedOutput, string(bytes))
		})
	}
}
