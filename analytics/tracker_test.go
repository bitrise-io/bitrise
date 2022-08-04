package analytics

import (
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/stretchr/testify/require"
)

func Test_mapStepResultToEvent(t *testing.T) {
	tests := []struct {
		name               string
		result             StepResult
		expectedEvent      string
		expectedExtraProps analytics.Properties
	}{
		{
			name: "Step succeeded",
			result: StepResult{
				Status: models.StepRunStatusCodeSuccess,
			},
			expectedEvent:      "step_finished",
			expectedExtraProps: analytics.Properties{"status": "successful"},
		},
		{
			name: "Step failed",
			result: StepResult{
				Status:       models.StepRunStatusCodeFailed,
				ErrorMessage: "msg",
			},
			expectedEvent: "step_finished",
			expectedExtraProps: analytics.Properties{
				"status":        "failed",
				"error_message": "msg",
			},
		},
		{
			name: "Step failed, skippable",
			result: StepResult{
				Status: models.StepRunStatusCodeFailedSkippable,
			},
			expectedEvent:      "step_finished",
			expectedExtraProps: analytics.Properties{"status": "failed"},
		},
		{
			name: "Step skipped",
			result: StepResult{
				Status: models.StepRunStatusCodeSkipped,
				Info: StepInfo{
					StepID:    "ID",
					Skippable: true,
				},
			},
			expectedEvent: "step_skipped",
			expectedExtraProps: analytics.Properties{
				"reason":    "build_failed",
				"skippable": true,
				"step_id":   "ID",
			},
		},
		{
			name: "Step skipped with run if",
			result: StepResult{
				Status: models.StepRunStatusCodeSkippedWithRunIf,
			},
			expectedEvent: "step_skipped",
			expectedExtraProps: analytics.Properties{
				"reason":    "run_if",
				"skippable": false,
			},
		},
		{
			name: "Step preparation failed",
			result: StepResult{
				Status:       models.StepRunStatusCodePreparationFailed,
				ErrorMessage: "msg",
			},
			expectedEvent: "step_preparation_failed",
			expectedExtraProps: analytics.Properties{
				"skippable":     false,
				"error_message": "msg",
			},
		},
		{
			name: "Step timeout",
			result: StepResult{
				Status:  models.StepRunStatusAbortedTimeout,
				Timeout: time.Second,
			},
			expectedEvent: "step_aborted",
			expectedExtraProps: analytics.Properties{
				"reason":  "timeout",
				"timeout": int64(1),
			},
		},
		{
			name: "Step timeout",
			result: StepResult{
				Status:          models.StepRunStatusAbortedNoOutputTimeout,
				NoOutputTimeout: time.Second,
			},
			expectedEvent: "step_aborted",
			expectedExtraProps: analytics.Properties{
				"reason":  "no_output_timeout",
				"timeout": int64(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualEvent, actualProps, err := mapStepResultToEvent(tt.result)

			require.NoError(t, err)
			require.Equal(t, tt.expectedEvent, actualEvent)
			require.Equal(t, tt.expectedExtraProps, actualProps)
		})
	}
}
