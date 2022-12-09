package models

import (
	"fmt"
	"strings"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// FormatVersion ...
	FormatVersion = "12"
)

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

// PipelineModel ...
type PipelineModel struct {
	Stages []StageListItemModel `json:"stages,omitempty" yaml:"stages,omitempty"`
}

// StageListItemModel ...
type StageListItemModel map[string]StageModel

// StageModel ...
type StageModel struct {
	Workflows []WorkflowListItemModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// WorkflowListItemModel ...
type WorkflowListItemModel map[string]WorkflowModel

// WorkflowModel ...
type WorkflowModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	BeforeRun    []string                            `json:"before_run,omitempty" yaml:"before_run,omitempty"`
	AfterRun     []string                            `json:"after_run,omitempty" yaml:"after_run,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps        []StepListItemModel                 `json:"steps,omitempty" yaml:"steps,omitempty"`
	Meta         map[string]interface{}              `json:"meta,omitempty" yaml:"meta,omitempty"`
}

// AppModel ...
type AppModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
}

// TriggerEventType ...
type TriggerEventType string

const (
	// TriggerEventTypeCodePush ...
	TriggerEventTypeCodePush TriggerEventType = "code-push"
	// TriggerEventTypePullRequest ...
	TriggerEventTypePullRequest TriggerEventType = "pull-request"
	// TriggerEventTypeTag ...
	TriggerEventTypeTag TriggerEventType = "tag"
	// TriggerEventTypeUnknown ...
	TriggerEventTypeUnknown TriggerEventType = "unknown"
)

// TriggerMapItemModel ...
type TriggerMapItemModel struct {
	PushBranch              string `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	PullRequestSourceBranch string `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch string `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	Tag                     string `json:"tag,omitempty" yaml:"tag,omitempty"`
	PipelineID              string `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	WorkflowID              string `json:"workflow,omitempty" yaml:"workflow,omitempty"`

	// deprecated
	Pattern              string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	IsPullRequestAllowed bool   `json:"is_pull_request_allowed,omitempty" yaml:"is_pull_request_allowed,omitempty"`
}

// TriggerMapModel ...
type TriggerMapModel []TriggerMapItemModel

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion        string `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string `json:"default_step_lib_source,omitempty" yaml:"default_step_lib_source,omitempty"`
	ProjectType          string `json:"project_type" yaml:"project_type"`
	//
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	//
	App        AppModel                 `json:"app,omitempty" yaml:"app,omitempty"`
	Meta       map[string]interface{}   `json:"meta,omitempty" yaml:"meta,omitempty"`
	TriggerMap TriggerMapModel          `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Pipelines  map[string]PipelineModel `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Stages     map[string]StageModel    `json:"stages,omitempty" yaml:"stages,omitempty"`
	Workflows  map[string]WorkflowModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// StepIDData ...
// structured representation of a composite-step-id
//	a composite step id is: step-lib-source::step-id@1.0.0
type StepIDData struct {
	// SteplibSource : steplib source uri, or in case of local path just "path", and in case of direct git url just "git"
	SteplibSource string
	// IDOrURI : ID if steplib is provided, URI if local step or in case a direct git url provided
	IDorURI string
	// Version : version in the steplib, or in case of a direct git step the tag-or-branch to use
	Version string
}

// BuildRunStartModel ...
type BuildRunStartModel struct {
	EventName   string    `json:"event_name" yaml:"event_name"`
	ProjectType string    `json:"project_type" yaml:"project_type"`
	StartTime   time.Time `json:"start_time" yaml:"start_time"`
}

// BuildRunResultsModel ...
type BuildRunResultsModel struct {
	WorkflowID           string                `json:"workflow_id" yaml:"workflow_id"`
	EventName            string                `json:"event_name" yaml:"event_name"`
	ProjectType          string                `json:"project_type" yaml:"project_type"`
	StartTime            time.Time             `json:"start_time" yaml:"start_time"`
	StepmanUpdates       map[string]int        `json:"stepman_updates" yaml:"stepman_updates"`
	SuccessSteps         []StepRunResultsModel `json:"success_steps" yaml:"success_steps"`
	FailedSteps          []StepRunResultsModel `json:"failed_steps" yaml:"failed_steps"`
	FailedSkippableSteps []StepRunResultsModel `json:"failed_skippable_steps" yaml:"failed_skippable_steps"`
	SkippedSteps         []StepRunResultsModel `json:"skipped_steps" yaml:"skipped_steps"`
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	StepInfo   stepmanModels.StepInfoModel `json:"step_info" yaml:"step_info"`
	StepInputs map[string]string           `json:"step_inputs" yaml:"step_inputs"`
	Status     StepRunStatus               `json:"status" yaml:"status"`
	Idx        int                         `json:"idx" yaml:"idx"`
	RunTime    time.Duration               `json:"run_time" yaml:"run_time"`
	StartTime  time.Time                   `json:"start_time" yaml:"start_time"`
	ErrorStr   string                      `json:"error_str" yaml:"error_str"`
	ExitCode   int                         `json:"exit_code" yaml:"exit_code"`
}

// StepError ...
type StepError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s StepRunResultsModel) StatusReasons() (string, []StepError) {
	switch s.Status {
	case StepRunStatusCodeSuccess:
		return "", nil
	case StepRunStatusCodeSkipped:
		return s.statusReason(), nil
	case StepRunStatusCodeSkippedWithRunIf:
		return s.statusReason(), nil
	case StepRunStatusCodeFailedSkippable:
		return s.statusReason(), s.error()
	case StepRunStatusCodeFailed:
		return "", s.error()
	case StepRunStatusCodePreparationFailed:
		return "", s.error()
	case StepRunStatusAbortedWithCustomTimeout:
		return "", s.error()
	case StepRunStatusAbortedWithNoOutputTimeout:
		return "", s.error()
	default:
		return "", nil
	}
}

func (s StepRunResultsModel) statusReason() string {
	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed,
		StepRunStatusAbortedWithCustomTimeout,
		StepRunStatusAbortedWithNoOutputTimeout:
		return ""
	case StepRunStatusCodeFailedSkippable:
		return `This Step failed, but it was marked as "is_skippable", so the build continued.`
	case StepRunStatusCodeSkipped:
		return `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`
	case StepRunStatusCodeSkippedWithRunIf:
		return fmt.Sprintf(`This Step was skipped, because its "run_if" expression evaluated to false.
The "run_if" expression was: %s`, *s.StepInfo.Step.RunIf)
	}

	return ""
}

func (s StepRunResultsModel) error() []StepError {
	message := ""

	switch s.Status {
	case StepRunStatusCodeSuccess,
		StepRunStatusCodeSkipped,
		StepRunStatusCodeSkippedWithRunIf:
		return nil
	case StepRunStatusCodeFailedSkippable,
		StepRunStatusCodeFailed,
		StepRunStatusCodePreparationFailed:
		message = s.ErrorStr
	case StepRunStatusAbortedWithCustomTimeout:
		message = fmt.Sprintf("This Step timed out after %s.", formatStatusReasonTimeInterval(*s.StepInfo.Step.Timeout))
	case StepRunStatusAbortedWithNoOutputTimeout:
		message = fmt.Sprintf("This Step failed, because it has not sent any output for %s.", formatStatusReasonTimeInterval(*s.StepInfo.Step.NoOutputTimeout))
	}

	return []StepError{{
		Code:    s.ExitCode,
		Message: message,
	}}
}

func formatStatusReasonTimeInterval(timeInterval int) string {
	var remaining int = timeInterval
	h := int(remaining / 3600)
	remaining = remaining - h*3600
	m := int(remaining / 60)
	remaining = remaining - m*60
	s := remaining

	var formattedTimeInterval = ""
	if h > 0 {
		formattedTimeInterval += fmt.Sprintf("%dh ", h)
	}

	if m > 0 {
		formattedTimeInterval += fmt.Sprintf("%dm ", m)
	}

	if s > 0 {
		formattedTimeInterval += fmt.Sprintf("%ds", s)
	}

	formattedTimeInterval = strings.TrimSpace(formattedTimeInterval)

	return formattedTimeInterval
}

// TestResultStepInfo ...
type TestResultStepInfo struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Title   string `json:"title" yaml:"title"`
	Number  int    `json:"number" yaml:"number"`
}
