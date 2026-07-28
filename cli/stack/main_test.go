package stack

import (
	"os"
	"testing"
	"time"

	extanalytics "github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/stepman/activator"
	"github.com/bitrise-io/stepman/toolkits"

	cliAnalytics "github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// TestMain installs a no-op analytics tracker: NewListCommand's RunE calls
// cmdutil.LogCommandParameters, which panics on the package-level tracker's
// zero value if nothing has called cmdutil.SetTracker (normally done once at
// real CLI startup, cli/cli.go).
func TestMain(m *testing.M) {
	cmdutil.SetTracker(noOpTracker{})
	os.Exit(m.Run())
}

type noOpTracker struct{}

func (noOpTracker) SendStepStartedEvent(extanalytics.Properties, cliAnalytics.StepInfo, time.Duration, map[string]interface{}, map[string]string) {
}
func (noOpTracker) SendStepFinishedEvent(extanalytics.Properties, cliAnalytics.StepResult) {}
func (noOpTracker) SendCLIWarning(string)                                                  {}
func (noOpTracker) SendWorkflowStarted(extanalytics.Properties, string, string)            {}
func (noOpTracker) SendWorkflowFinished(extanalytics.Properties, bool)                     {}
func (noOpTracker) SendCommandInfo(string, string, []string)                               {}
func (noOpTracker) SendToolSetupEvent(string, provider.ToolRequest, provider.ToolInstallResult, bool, time.Duration) {
}
func (noOpTracker) SendStepActivationEvent(activator.ActivationType, string, bool, time.Duration, bool) {
}
func (noOpTracker) SendToolkitPrepareEvent(string, string, string, string, toolkits.PrepareForStepRunResult, error) {
}
func (noOpTracker) Wait()            {}
func (noOpTracker) IsTracking() bool { return false }
