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
	Version = "1.0.0"
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

// TriggerMapItemModel ...
type TriggerMapItemModel struct {
	Pattern              string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	IsPullRequestAllowed bool   `json:"is_pull_request_allowed,omitempty" yaml:"is_pull_request_allowed,omitempty"`
	WorkflowID           string `json:"workflow,omitempty" yaml:"workflow,omitempty"`
}

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion        string `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string `json:"default_step_lib_source,omitempty" yaml:"default_step_lib_source,omitempty"`
	//
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	//
	App        AppModel                 `json:"app,omitempty" yaml:"app,omitempty"`
	TriggerMap []TriggerMapItemModel    `json:"trigger_map,omitempty" yaml:"trigger_map,omitempty"`
	Workflows  map[string]WorkflowModel `json:"workflows,omitempty" yaml:"workflows,omitempty"`
}

// StepIDData ...
// structured representation of a composite-step-id
//  a composite step id is: step-lib-source::step-id@1.0.0
type StepIDData struct {
	// IDOrURI : ID if steplib is provided, URI if local step or in case a direct git url provided
	IDorURI string
	// Version : version in the steplib, or in case of a direct git step the tag-or-branch to use
	Version string
	// SteplibSource : steplib source uri, or in case of local path just "path", and in case of direct git url just "git"
	SteplibSource string
}

// BuildRunResultsModel ...
type BuildRunResultsModel struct {
	StartTime            time.Time
	StepmanUpdates       map[string]int
	SuccessSteps         []StepRunResultsModel
	FailedSteps          []StepRunResultsModel
	FailedSkippableSteps []StepRunResultsModel
	SkippedSteps         []StepRunResultsModel
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	StepName string
	Status   int
	Idx      int
	RunTime  time.Duration
	Error    error
	ExitCode int
}
