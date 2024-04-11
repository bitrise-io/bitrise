package log

import (
	"encoding/json"
	"testing"

	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/assert"
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
				ExecutionID: "ExecutionId",
				Position:    0,
				Title:       "Title",
				ID:          "Id",
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
				ExecutionID: "ExecutionId",
				Position:    0,
				Title:       "Title",
				ID:          "Id",
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

func TestStepFinishedEventSerialisesToTheExpectedJsonMessage(t *testing.T) {
	tests := []struct {
		name           string
		params         StepFinishedParams
		expectedOutput string
	}{
		{
			name: "Fields are serialising correctly",
			params: StepFinishedParams{
				ExecutionID:   "ExecutionId",
				Status:        models.StepRunStatusCodeFailed.String(),
				StatusReason:  "StatusReason",
				Title:         "Title",
				RunTime:       1234567890,
				SupportURL:    "SupportURL",
				SourceCodeURL: "SourceCodeURL",
				Errors: []models.StepError{
					{
						Code:    2,
						Message: "Message",
					},
				},
				Update: &StepUpdate{
					OriginalVersion: "OriginalVersion",
					ResolvedVersion: "ResolvedVersion",
					LatestVersion:   "LatestVersion",
					ReleasesURL:     "ReleasesURL",
				},
				Deprecation: &StepDeprecation{
					RemovalDate: "RemovalDate",
					Note:        "Note",
				},
				LastStep: false,
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"status\":\"failed\",\"status_reason\":\"StatusReason\",\"title\":\"Title\",\"run_time_in_ms\":1234567890,\"support_url\":\"SupportURL\",\"source_code_url\":\"SourceCodeURL\",\"errors\":[{\"code\":2,\"message\":\"Message\"}],\"update_available\":{\"original_version\":\"OriginalVersion\",\"resolved_version\":\"ResolvedVersion\",\"latest_version\":\"LatestVersion\",\"release_notes\":\"ReleasesURL\"},\"deprecation\":{\"removal_date\":\"RemovalDate\",\"note\":\"Note\"},\"last_step\":false}",
		},
		{
			name: "Optional fields are omitted when empty",
			params: StepFinishedParams{
				ExecutionID:   "ExecutionId",
				Status:        models.StepRunStatusCodeSkipped.String(),
				StatusReason:  "",
				Title:         "Title",
				RunTime:       1234567890,
				SupportURL:    "SupportURL",
				SourceCodeURL: "SourceCodeURL",
				Errors:        []models.StepError{},
				Update:        nil,
				Deprecation:   nil,
				LastStep:      true,
			},
			expectedOutput: "{\"uuid\":\"ExecutionId\",\"status\":\"skipped\",\"title\":\"Title\",\"run_time_in_ms\":1234567890,\"support_url\":\"SupportURL\",\"source_code_url\":\"SourceCodeURL\",\"last_step\":true}",
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
