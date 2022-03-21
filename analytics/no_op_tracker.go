package analytics

import (
	"path"

	models2 "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
)

type analyticsPluginConfigModel struct {
	IsAnalyticsDisabled bool `yaml:"is_analytics_disabled"`
}

// NoOpTracker ...
type NoOpTracker struct {
}

// SendWorkflowStarted ...
func (n NoOpTracker) SendWorkflowStarted(analytics.Properties, string) {
}

// SendWorkflowFinished ...
func (n NoOpTracker) SendWorkflowFinished(analytics.Properties, bool) {
}

// SendStepStartedEvent ...
func (n NoOpTracker) SendStepStartedEvent(analytics.Properties, models.StepInfoModel, map[string]interface{}) {
}

// SendStepFinishedEvent ...
func (n NoOpTracker) SendStepFinishedEvent(analytics.Properties, models2.StepRunResultsModel) {
}

// Wait ...
func (n NoOpTracker) Wait() {

}

func isAnalyticsDisabled() bool {
	config := analyticsPluginConfigModel{}
	configPth := path.Join(plugins.GetPluginDataDir("analytics"), "config.yml")
	if exist, err := pathutil.IsPathExists(configPth); err == nil && exist {
		bytes, err := fileutil.ReadBytesFromFile(configPth)
		if err == nil {
			err := yaml.Unmarshal(bytes, &config)
			if err == nil {
				return config.IsAnalyticsDisabled
			}
		}
	}
	return false
}
