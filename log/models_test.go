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
			name: "Every field is serialised",
			params: StepStartedParams{
				ExecutionId: "ExecutionId",
				Position:    0,
				Title:       "Title",
				Id:          "Id",
				Version:     "Version",
				Collection:  "Collection",
				Toolkit:     "Toolkit",
				StartTime:   "StartTime",
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"idx\":0,\"title\":\"Title\",\"id\":\"Id\",\"version\":\"Version\",\"collection\":\"Collection\",\"toolkit\":\"Toolkit\",\"start_time\":\"StartTime\"}",
		},
		{
			name: "Empty fields are kept",
			params: StepStartedParams{
				ExecutionId: "ExecutionId",
				Position:    0,
				Title:       "Title",
				Id:          "Id",
				Version:     "",
				Collection:  "Collection",
				Toolkit:     "",
				StartTime:   "StartTime",
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"idx\":0,\"title\":\"Title\",\"id\":\"Id\",\"version\":\"\",\"collection\":\"Collection\",\"toolkit\":\"\",\"start_time\":\"StartTime\"}",
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
