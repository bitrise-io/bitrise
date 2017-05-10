package models

import (
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// StepRunStatusCodeSuccess ...
	StepRunStatusCodeSuccess = 0
	// StepRunStatusCodeFailed ...
	StepRunStatusCodeFailed = 1
	// StepRunStatusCodeFailedSkippable ...
	StepRunStatusCodeFailedSkippable = 2
	// StepRunStatusCodeSkipped ...
	StepRunStatusCodeSkipped = 3
	// StepRunStatusCodeSkippedWithRunIf ...
	StepRunStatusCodeSkippedWithRunIf = 4

	// Version ...
	Version = "3"
)

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

// WorkflowModel ...
type WorkflowModel struct {
	Title        string                              `json:"title,omitempty" yaml:"title,omitempty"`
	Summary      string                              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                              `json:"description,omitempty" yaml:"description,omitempty"`
	BeforeRun    []string                            `json:"before_run,omitempty" yaml:"before_run,omitempty"`
	AfterRun     []string                            `json:"after_run,omitempty" yaml:"after_run,omitempty"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs,omitempty" yaml:"envs,omitempty"`
	Steps        []StepListItemModel                 `json:"steps,omitempty" yaml:"steps,omitempty"`
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
	TriggerEventTypeUnknown TriggerEventType = "unkown"
)

// TriggerMapItemModel ...
type TriggerMapItemModel struct {
	PushBranch              string `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	PullRequestSourceBranch string `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch string `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	Tag                     string `json:"tag,omitempty" yaml:"tag,omitempty"`
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
	TriggerMap TriggerMapModel          `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Workflows  map[string]WorkflowModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// StepIDData ...
// structured representation of a composite-step-id
//  a composite step id is: step-lib-source::step-id@1.0.0
type StepIDData struct {
	// SteplibSource : steplib source uri, or in case of local path just "path", and in case of direct git url just "git"
	SteplibSource string
	// IDOrURI : ID if steplib is provided, URI if local step or in case a direct git url provided
	IDorURI string
	// Version : version in the steplib, or in case of a direct git step the tag-or-branch to use
	Version string
}

// BuildRunResultsModel ...
type BuildRunResultsModel struct {
	StartTime            time.Time             `json:"start_time" yaml:"start_time"`
	StepmanUpdates       map[string]int        `json:"stepman_updates" yaml:"stepman_updates"`
	SuccessSteps         []StepRunResultsModel `json:"success_steps" yaml:"success_steps"`
	FailedSteps          []StepRunResultsModel `json:"failed_steps" yaml:"failed_steps"`
	FailedSkippableSteps []StepRunResultsModel `json:"failed_skippable_steps" yaml:"failed_skippable_steps"`
	SkippedSteps         []StepRunResultsModel `json:"skipped_steps" yaml:"skipped_steps"`
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	StepInfo stepmanModels.StepInfoModel `json:"step_info" yaml:"step_info"`
	Status   int                         `json:"status" yaml:"status"`
	Idx      int                         `json:"idx" yaml:"idx"`
	RunTime  time.Duration               `json:"run_time" yaml:"run_time"`
	ErrorStr string                      `json:"error_str" yaml:"error_str"`
	ExitCode int                         `json:"exit_code" yaml:"exit_code"`
}
