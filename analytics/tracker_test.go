package analytics

import (
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/stepman/activator"
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
				Status:  models.StepRunStatusCodeSuccess,
				Runtime: 30 * time.Second,
				Info:    StepInfo{StepID: "ID"},
			},
			expectedEvent:      "step_finished",
			expectedExtraProps: analytics.Properties{"status": "successful", "runtime": int64(30), "step_id": "ID"},
		},
		{
			name: "Step failed",
			result: StepResult{
				Status:       models.StepRunStatusCodeFailed,
				ErrorMessage: "msg",
				Info:         StepInfo{StepID: "ID"},
			},
			expectedEvent: "step_finished",
			expectedExtraProps: analytics.Properties{
				"status":        "failed",
				"error_message": "msg",
				"runtime":       int64(0),
				"step_id":       "ID",
			},
		},
		{
			name: "Step failed, skippable",
			result: StepResult{
				Status: models.StepRunStatusCodeFailedSkippable,
			},
			expectedEvent:      "step_finished",
			expectedExtraProps: analytics.Properties{"status": "failed", "runtime": int64(0)},
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
				"runtime":   int64(0),
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
				"runtime":   int64(0),
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
				"runtime":       int64(0),
			},
		},
		{
			name: "Step timeout",
			result: StepResult{
				Status:  models.StepRunStatusAbortedWithCustomTimeout,
				Timeout: time.Second,
			},
			expectedEvent: "step_aborted",
			expectedExtraProps: analytics.Properties{
				"reason":  "timeout",
				"timeout": int64(1),
				"runtime": int64(0),
			},
		},
		{
			name: "Step timeout",
			result: StepResult{
				Status:          models.StepRunStatusAbortedWithNoOutputTimeout,
				NoOutputTimeout: time.Second,
			},
			expectedEvent: "step_aborted",
			expectedExtraProps: analytics.Properties{
				"reason":  "no_output_timeout",
				"timeout": int64(1),
				"runtime": int64(0),
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

// capturingAnalyticsTracker records what the tracker hands to the analytics client.
type capturingAnalyticsTracker struct {
	events []capturedEvent
}

type capturedEvent struct {
	name  string
	props analytics.Properties
}

func (c *capturingAnalyticsTracker) Enqueue(eventName string, properties ...analytics.Properties) {
	merged := analytics.Properties{}
	for _, p := range properties {
		merged = merged.Merge(p)
	}
	c.events = append(c.events, capturedEvent{name: eventName, props: merged})
}

func (c *capturingAnalyticsTracker) Wait()            {}
func (c *capturingAnalyticsTracker) IsTracking() bool { return true }

func Test_SendStepActivationEvent(t *testing.T) {
	tests := []struct {
		name                string
		activationType      activator.ActivationType
		inventorySource     activator.ActivationInventorySource
		isSuccessful        bool
		wantInventorySource any // nil means the key must be absent
	}{
		{
			name:                "StepLib API activation reports steplib_api",
			activationType:      activator.ActivationTypeSteplibSource,
			inventorySource:     activator.ActivationInventorySourceSteplibAPI,
			isSuccessful:        true,
			wantInventorySource: "steplib_api",
		},
		{
			name:                "Legacy activation reports steplib",
			activationType:      activator.ActivationTypeSteplibSource,
			inventorySource:     activator.ActivationInventorySourceSteplib,
			isSuccessful:        true,
			wantInventorySource: "steplib",
		},
		{
			// The tracker must not gate the inventory source on success the way it
			// gates did_steplib_update below: a failed activation is still
			// attributable to a path, which is the point of reporting it. That
			// stepman populates the field on its error paths is covered upstream,
			// in activator.TestActivateSteplibRefStep.
			name:                "Failed activation still reports the inventory source",
			activationType:      activator.ActivationTypeSteplibSource,
			inventorySource:     activator.ActivationInventorySourceSteplibAPI,
			isSuccessful:        false,
			wantInventorySource: "steplib_api",
		},
		{
			name:                "Path ref omits the inventory source",
			activationType:      activator.ActivationTypePathRef,
			inventorySource:     activator.ActivationInventorySourceNone,
			isSuccessful:        true,
			wantInventorySource: nil,
		},
		{
			name:                "Git ref omits the inventory source",
			activationType:      activator.ActivationTypeGitRef,
			inventorySource:     activator.ActivationInventorySourceNone,
			isSuccessful:        true,
			wantInventorySource: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(DisabledEnvKey, "")

			envRepository := env.NewRepository()
			analyticsTracker := &capturingAnalyticsTracker{}
			logger := log.NewUtilsLogAdapter()
			subject := NewTracker(analyticsTracker, envRepository, NewStateChecker(envRepository), &logger)

			subject.SendStepActivationEvent(
				"step-execution-id",
				tt.activationType,
				tt.inventorySource,
				"git-clone@8.5.0",
				tt.isSuccessful,
				2*time.Second,
				true,
			)

			require.Len(t, analyticsTracker.events, 1)
			event := analyticsTracker.events[0]
			require.Equal(t, "cli_step_activation", event.name)

			// step_execution_id makes this event joinable with cli_toolkit_prepare.
			require.Equal(t, "step-execution-id", event.props["step_execution_id"])
			require.Equal(t, tt.activationType, event.props["activation_type"])
			require.Equal(t, int64(2000), event.props["duration_ms"])
			require.Equal(t, tt.isSuccessful, event.props["is_successful"])

			if tt.wantInventorySource == nil {
				require.NotContains(t, event.props, "inventory_source")
			} else {
				require.Equal(t, tt.wantInventorySource, event.props["inventory_source"])
			}

			// did_steplib_update is reported only for successful activations, so the
			// expectation follows isSuccessful rather than being a separate column.
			if tt.isSuccessful {
				require.Equal(t, true, event.props["did_steplib_update"])
			} else {
				require.NotContains(t, event.props, "did_steplib_update")
			}
		})
	}
}
