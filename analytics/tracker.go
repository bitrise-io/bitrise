package analytics

import (
	"encoding/json"
	"os"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
	stepmanModels "github.com/bitrise-io/stepman/models"
	log "github.com/sirupsen/logrus"
)

// Tracker ...
type Tracker interface {
	SendWorkflowStarted(properties analytics.Properties, title string)
	SendWorkflowFinished(properties analytics.Properties, failed bool)
	SendStepStartedEvent(properties analytics.Properties, infoModel stepmanModels.StepInfoModel, inputs map[string]interface{})
	SendStepFinishedEvent(properties analytics.Properties, results models.StepRunResultsModel)
	Wait()
}

type tracker struct {
	tracker       analytics.Tracker
	envRepository env.Repository
}

// NewTracker ...
func NewTracker() Tracker {
	if isAnalyticsDisabled() {
		return NoOpTracker{}
	}
	return tracker{
		tracker:       analytics.NewDefaultTracker(),
		envRepository: env.NewRepository(),
	}
}

// SendWorkflowStarted ...
func (t tracker) SendWorkflowStarted(properties analytics.Properties, title string) {
	isCI := t.envRepository.Get(configs.CIModeEnvKey) == "true"
	isPR := t.envRepository.Get(configs.PRModeEnvKey) == "true"
	isDebug := t.envRepository.Get(configs.DebugModeEnvKey) == "true"
	isSecretFiltering := t.envRepository.Get(configs.IsSecretFilteringKey) == "true"
	isSecretEnvsFiltering := t.envRepository.Get(configs.IsSecretEnvsFilteringKey) == "true"
	stateProperties := analytics.Properties{
		"workflow_name":         title,
		"ci_mode":               isCI,
		"pr_mode":               isPR,
		"debug_mode":            isDebug,
		"secret_filtering":      isSecretFiltering,
		"secret_envs_filtering": isSecretEnvsFiltering,
	}
	stateProperties.AppendIfNotEmpty("build_slug", t.envRepository.Get("BITRISE_BUILD_SLUG"))
	stateProperties.AppendIfNotEmpty("parent_step_unique_id", t.envRepository.Get("BITRISE_STEP_UNIQUE_ID"))
	stateProperties.AppendIfNotEmpty("log_level", t.envRepository.Get(configs.LogLevelEnvKey))
	currentVersionMap, err := version.ToolVersionMap(os.Args[0])
	if err == nil {
		bitriseVersion := currentVersionMap["bitrise"]
		envmanVersion := currentVersionMap["envman"]
		stepmanVersion := currentVersionMap["stepman"]
		stateProperties.AppendIfNotEmpty("cli_version", bitriseVersion.String())
		stateProperties.AppendIfNotEmpty("envman_version", envmanVersion.String())
		stateProperties.AppendIfNotEmpty("stepman_version", stepmanVersion.String())
	} else {
		log.Debugf("Couldn't get tool versions: %s", err.Error())
	}
	t.tracker.Enqueue("workflow_started", properties, stateProperties)
}

// SendWorkflowFinished ...
func (t tracker) SendWorkflowFinished(properties analytics.Properties, failed bool) {
	var statusMessage string
	if failed {
		statusMessage = "failed"
	} else {
		statusMessage = "successful"
	}
	t.tracker.Enqueue("workflow_finished", properties, analytics.Properties{"status": statusMessage})
}

// SendStepStartedEvent ...
func (t tracker) SendStepStartedEvent(properties analytics.Properties, infoModel stepmanModels.StepInfoModel, inputs map[string]interface{}) {
	inputBytes, err := json.Marshal(inputs)
	if err != nil {
		log.Errorf("Failed to marshal inputs: %s", err)
	}
	t.tracker.Enqueue("step_started", properties, prepareStartProperties(infoModel), analytics.Properties{"inputs": string(inputBytes)})
}

// SendStepFinishedEvent ...
func (t tracker) SendStepFinishedEvent(properties analytics.Properties, results models.StepRunResultsModel) {
	switch results.Status {
	case models.StepRunStatusCodeSuccess:
		t.tracker.Enqueue("step_finished", properties, analytics.Properties{"status": "successful"})
		break
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodeFailedSkippable:
		failedProperties := analytics.Properties{"status": "failed"}
		failedProperties.AppendIfNotEmpty("error_message", results.ErrorStr)
		t.tracker.Enqueue("step_finished", properties, failedProperties)
		break
	case models.StepRunStatusCodePreparationFailed:
		failedProperties := prepareStartProperties(results.StepInfo)
		failedProperties.AppendIfNotEmpty("error_message", results.ErrorStr)
		t.tracker.Enqueue("step_preparation_failed", properties, failedProperties)
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		startProperties := prepareStartProperties(results.StepInfo)
		if results.Status == models.StepRunStatusCodeSkipped {
			startProperties["reason"] = "build_failed"
		} else {
			startProperties["reason"] = "run_if"
		}
		t.tracker.Enqueue("step_skipped", properties, startProperties)
	default:
		panic("Unknown step status code")
	}
}

func prepareStartProperties(infoModel stepmanModels.StepInfoModel) analytics.Properties {
	properties := analytics.Properties{}
	properties.AppendIfNotEmpty("step_id", infoModel.ID)
	properties.AppendIfNotEmpty("step_title", pointers.StringWithDefault(infoModel.Step.Title, ""))
	properties.AppendIfNotEmpty("step_version", infoModel.Version)
	properties.AppendIfNotEmpty("step_source", pointers.StringWithDefault(infoModel.Step.SourceCodeURL, ""))
	properties["skippable"] = pointers.BoolWithDefault(infoModel.Step.IsSkippable, false)
	return properties
}

// Wait ...
func (t tracker) Wait() {
	t.tracker.Wait()
}
