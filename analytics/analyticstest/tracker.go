// Package analyticstest provides a no-op analytics.Tracker for tests —
// either wired via cmdutil.SetTracker (command packages whose RunE calls
// cmdutil.LogCommandParameters) or passed directly to a constructor
// expecting a Tracker (e.g. NewWorkflowRunner).
package analyticstest

import (
	"time"

	extanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/stepman/activator"
	"github.com/bitrise-io/stepman/toolkits"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// NoOpTracker implements analytics.Tracker with no-op methods.
type NoOpTracker struct{}

func (NoOpTracker) SendStepStartedEvent(extanalytics.Properties, analytics.StepInfo, time.Duration, map[string]interface{}, map[string]string) {
}
func (NoOpTracker) SendStepFinishedEvent(extanalytics.Properties, analytics.StepResult) {}
func (NoOpTracker) SendCLIWarning(string)                                               {}
func (NoOpTracker) SendWorkflowStarted(extanalytics.Properties, string, string)         {}
func (NoOpTracker) SendWorkflowFinished(extanalytics.Properties, bool)                  {}
func (NoOpTracker) SendCommandInfo(string, string, []string)                            {}
func (NoOpTracker) SendToolSetupEvent(string, provider.ToolRequest, provider.ToolInstallResult, bool, time.Duration) {
}
func (NoOpTracker) SendStepActivationEvent(string, activator.ActivationType, activator.ActivationInventorySource, string, bool, time.Duration, bool) {
}
func (NoOpTracker) SendToolkitPrepareEvent(string, string, string, string, toolkits.PrepareForStepRunResult, error) {
}
func (NoOpTracker) Wait()            {}
func (NoOpTracker) IsTracking() bool { return false }
