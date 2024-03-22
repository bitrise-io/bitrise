package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
	utilslog "github.com/bitrise-io/go-utils/v2/log"
)

const (
	BuildExecutionID    = "build_execution_id"
	WorkflowExecutionID = "workflow_execution_id"
	StepExecutionID     = "step_execution_id"

	workflowStartedEventName       = "workflow_started"
	workflowFinishedEventName      = "workflow_finished"
	stepStartedEventName           = "step_started"
	stepFinishedEventName          = "step_finished"
	stepAbortedEventName           = "step_aborted"
	stepPreparationFailedEventName = "step_preparation_failed"
	stepSkippedEventName           = "step_skipped"
	cliWarningEventName            = "cli_warning"
	toolVersionSnapshotEventName   = "tool_version_snapshot"

	workflowNameProperty          = "workflow_name"
	workflowTitleProperty         = "workflow_title"
	ciModeProperty                = "ci_mode"
	prModeProperty                = "pr_mode"
	debugModeProperty             = "debug_mode"
	secretFilteringProperty       = "secret_filtering"
	secretEnvsFilteringProperty   = "secret_envs_filtering"
	buildSlugProperty             = "build_slug"
	appSlugProperty               = "app_slug"
	parentStepExecutionIDProperty = "parent_step_execution_id"
	cliVersionProperty            = "cli_version"
	envmanVersionProperty         = "envman_version"
	stepmanVersionProperty        = "stepman_version"
	statusProperty                = "status"
	inputsProperty                = "inputs"
	errorMessageProperty          = "error_message"
	reasonProperty                = "reason"
	messageProperty               = "message"
	stepIDProperty                = "step_id"
	stepTitleProperty             = "step_title"
	stepVersionProperty           = "step_version"
	stepSourceProperty            = "step_source"
	skippableProperty             = "skippable"
	timeoutProperty               = "timeout"
	runTimeProperty               = "runtime"
	osProperty                    = "os"
	stackRevIDProperty            = "stack_rev_id"
	snapshotProperty              = "snapshot"
	toolVersionsProperty          = "tool_versions"

	failedValue                    = "failed"
	successfulValue                = "successful"
	buildFailedValue               = "build_failed"
	runIfValue                     = "run_if"
	customTimeoutValue             = "timeout"
	noOutputTimeoutValue           = "no_output_timeout"
	ToolSnapshotEndOfWorkflowValue = "end_of_workflow"

	buildSlugEnvKey       = "BITRISE_BUILD_SLUG"
	appSlugEnvKey         = "BITRISE_APP_SLUG"
	StepExecutionIDEnvKey = "BITRISE_STEP_EXECUTION_ID"
	stackRevIDKey         = "BITRISE_STACK_REV_ID"
	macStackRevIDKey      = "BITRISE_OSX_STACK_REV_ID"

	bitriseVersionKey = "bitrise"
	envmanVersionKey  = "envman"
	stepmanVersionKey = "stepman"
)

type Input struct {
	Value         interface{} `json:"value"`
	OriginalValue string      `json:"original_value,omitempty"`
}

type StepInfo struct {
	StepID      string
	StepTitle   string
	StepVersion string
	StepSource  string
	Skippable   bool
}

type StepResult struct {
	Info                     StepInfo
	Status                   models.StepRunStatus
	ErrorMessage             string
	Timeout, NoOutputTimeout time.Duration
	Runtime                  time.Duration
}

type Tracker interface {
	SendWorkflowStarted(properties analytics.Properties, name string, title string)
	SendWorkflowFinished(properties analytics.Properties, failed bool)
	SendStepStartedEvent(properties analytics.Properties, info StepInfo, expandedInputs map[string]interface{}, originalInputs map[string]string)
	SendStepFinishedEvent(properties analytics.Properties, result StepResult)
	SendCLIWarning(message string)
	SendToolVersionSnapshot(toolVersions, snapshotType string)
	Wait()
}

type tracker struct {
	tracker       analytics.Tracker
	envRepository env.Repository
	stateChecker  StateChecker
	logger        utilslog.Logger
}

func NewTracker(analyticsTracker analytics.Tracker, envRepository env.Repository, stateChecker StateChecker, logger utilslog.Logger) Tracker {
	return tracker{tracker: analyticsTracker, envRepository: envRepository, stateChecker: stateChecker, logger: logger}
}

func NewDefaultTracker() Tracker {
	envRepository := env.NewRepository()
	stateChecker := NewStateChecker(envRepository)

	logger := log.NewUtilsLogAdapter()
	tracker := analytics.NewDefaultTracker(&logger)
	if stateChecker.UseAsync() {
		tracker = analytics.NewDefaultTracker(&logger)
	}

	return NewTracker(tracker, envRepository, stateChecker, &logger)
}

// SendWorkflowStarted sends `workflow_started` events. `parent_step_execution_id` can be used to filter those
// Bitrise CLI events that were started as part of a step (like script).
func (t tracker) SendWorkflowStarted(properties analytics.Properties, name string, title string) {
	if !t.stateChecker.Enabled() {
		return
	}

	isCI := t.envRepository.Get(configs.CIModeEnvKey) == "true"
	isPR := t.envRepository.Get(configs.PRModeEnvKey) == "true"
	isDebug := t.envRepository.Get(configs.DebugModeEnvKey) == "true"
	isSecretFiltering := t.envRepository.Get(configs.IsSecretFilteringKey) == "true"
	isSecretEnvsFiltering := t.envRepository.Get(configs.IsSecretEnvsFilteringKey) == "true"
	buildSlug := t.envRepository.Get(buildSlugEnvKey)
	appSlug := t.envRepository.Get(appSlugEnvKey)
	parentStepExecutionID := t.envRepository.Get(StepExecutionIDEnvKey)

	var bitriseVersion string
	var envmanVersion string
	var stepmanVersion string
	currentVersionMap, err := version.ToolVersionMap(os.Args[0])
	if err == nil {
		if bv, ok := currentVersionMap[bitriseVersionKey]; ok {
			bitriseVersion = bv.String()
		}
		if ev, ok := currentVersionMap[envmanVersionKey]; ok {
			envmanVersion = ev.String()
		}
		if sv, ok := currentVersionMap[stepmanVersionKey]; ok {
			stepmanVersion = sv.String()
		}
	} else {
		t.SendCLIWarning(fmt.Sprintf("Couldn't get tool versions: %s", err.Error()))
	}

	stateProperties := analytics.Properties{
		workflowNameProperty:        name,
		ciModeProperty:              isCI,
		prModeProperty:              isPR,
		debugModeProperty:           isDebug,
		secretFilteringProperty:     isSecretFiltering,
		secretEnvsFilteringProperty: isSecretEnvsFiltering,
	}
	if name != title && title != "" {
		stateProperties[workflowTitleProperty] = title
	}
	stateProperties.AppendIfNotEmpty(buildSlugProperty, buildSlug)
	stateProperties.AppendIfNotEmpty(appSlugProperty, appSlug)
	stateProperties.AppendIfNotEmpty(parentStepExecutionIDProperty, parentStepExecutionID)
	stateProperties.AppendIfNotEmpty(cliVersionProperty, bitriseVersion)
	stateProperties.AppendIfNotEmpty(envmanVersionProperty, envmanVersion)
	stateProperties.AppendIfNotEmpty(stepmanVersionProperty, stepmanVersion)

	t.tracker.Enqueue(workflowStartedEventName, properties, stateProperties)
}

func (t tracker) SendWorkflowFinished(properties analytics.Properties, failed bool) {
	if !t.stateChecker.Enabled() {
		return
	}

	var statusMessage string
	if failed {
		statusMessage = failedValue
	} else {
		statusMessage = successfulValue
	}

	t.tracker.Enqueue(workflowFinishedEventName, properties, analytics.Properties{statusProperty: statusMessage})
}

func (t tracker) SendToolVersionSnapshot(toolVersions, snapshotType string) {
	if !t.stateChecker.Enabled() {
		return
	}

	stackRevID := t.envRepository.Get(stackRevIDKey)
	if stackRevID == "" {
		// Legacy
		stackRevID = t.envRepository.Get(macStackRevIDKey)
	}

	properties := analytics.Properties{
		ciModeProperty:       t.envRepository.Get(configs.CIModeEnvKey) == "true",
		osProperty:           runtime.GOOS,
		stackRevIDProperty:   stackRevID,
		snapshotProperty:     snapshotType,
		toolVersionsProperty: toolVersions,
	}

	t.tracker.Enqueue(toolVersionSnapshotEventName, properties)
}

func (t tracker) SendStepStartedEvent(properties analytics.Properties, info StepInfo, expandedInputs map[string]interface{}, originalInputs map[string]string) {
	if !t.stateChecker.Enabled() {
		return
	}

	extraProperties := []analytics.Properties{properties, prepareStartProperties(info)}
	if len(expandedInputs) > 0 {
		inputMap := map[string]Input{}
		for k, v := range expandedInputs {
			inputMap[k] = Input{
				Value:         v,
				OriginalValue: originalInputs[k],
			}

		}
		inputBytes, err := json.Marshal(inputMap)
		if err != nil {
			t.SendCLIWarning(fmt.Sprintf("Failed to marshal inputs: %s", err.Error()))
		} else {
			extraProperties = append(extraProperties, analytics.Properties{inputsProperty: string(inputBytes)})
		}
	}

	t.tracker.Enqueue(stepStartedEventName, extraProperties...)
}

func (t tracker) SendStepFinishedEvent(properties analytics.Properties, result StepResult) {
	if !t.stateChecker.Enabled() {
		return
	}

	eventName, extraProperties, err := mapStepResultToEvent(result)
	if err != nil {
		t.SendCLIWarning(err.Error())
	}

	t.tracker.Enqueue(eventName, properties, extraProperties)
}

func (t tracker) SendCLIWarning(message string) {
	if !t.stateChecker.Enabled() {
		return
	}

	t.tracker.Enqueue(cliWarningEventName, analytics.Properties{messageProperty: message})
}

func (t tracker) Wait() {
	t.tracker.Wait()
}

func prepareStartProperties(info StepInfo) analytics.Properties {
	properties := analytics.Properties{}
	properties.AppendIfNotEmpty(stepIDProperty, info.StepID)
	properties.AppendIfNotEmpty(stepTitleProperty, info.StepTitle)
	properties.AppendIfNotEmpty(stepVersionProperty, info.StepVersion)
	properties.AppendIfNotEmpty(stepSourceProperty, info.StepSource)
	properties[skippableProperty] = info.Skippable

	return properties
}

func mapStepResultToEvent(result StepResult) (string, analytics.Properties, error) {
	var (
		eventName       string
		extraProperties analytics.Properties
	)

	switch result.Status {
	case models.StepRunStatusCodeSuccess:
		eventName = stepFinishedEventName
		extraProperties = analytics.Properties{statusProperty: successfulValue}
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodeFailedSkippable:
		eventName = stepFinishedEventName
		extraProperties = analytics.Properties{statusProperty: failedValue}
		extraProperties.AppendIfNotEmpty(errorMessageProperty, result.ErrorMessage)
	case models.StepRunStatusAbortedWithCustomTimeout:
		eventName = stepAbortedEventName
		extraProperties = analytics.Properties{reasonProperty: customTimeoutValue}

		if result.Timeout >= 0 {
			extraProperties[timeoutProperty] = int64(result.Timeout.Seconds())
		}
	case models.StepRunStatusAbortedWithNoOutputTimeout:
		eventName = stepAbortedEventName
		extraProperties = analytics.Properties{reasonProperty: noOutputTimeoutValue}

		if result.NoOutputTimeout >= 0 {
			extraProperties[timeoutProperty] = int64(result.NoOutputTimeout.Seconds())
		}
	case models.StepRunStatusCodePreparationFailed:
		eventName = stepPreparationFailedEventName
		extraProperties = prepareStartProperties(result.Info)
		extraProperties.AppendIfNotEmpty(errorMessageProperty, result.ErrorMessage)
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		eventName = stepSkippedEventName
		extraProperties = prepareStartProperties(result.Info)

		if result.Status == models.StepRunStatusCodeSkipped {
			extraProperties[reasonProperty] = buildFailedValue
		} else {
			extraProperties[reasonProperty] = runIfValue
		}
	default:
		return "", analytics.Properties{}, fmt.Errorf("Unknown step status code: %d", result.Status)
	}

	extraProperties[runTimeProperty] = int64(result.Runtime.Seconds())

	return eventName, extraProperties, nil
}
