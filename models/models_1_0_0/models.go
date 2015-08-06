package models

import (
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// -------------------
// --- bitrise models

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion        string                   `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string                   `json:"default_step_lib_source" yaml:"default_step_lib_source"`
	App                  AppModel                 `json:"app" yaml:"app"`
	Workflows            map[string]WorkflowModel `json:"workflows" yaml:"workflows"`
}

// AppModel ...
type AppModel struct {
	Environments []envmanModels.EnvironmentItemModel `json:"envs" yaml:"envs"`
}

// WorkflowModel ...
type WorkflowModel struct {
	Title        string                              `json:"title" yaml:"title"`
	Summary      string                              `json:"summary" yaml:"summary"`
	BeforeRun    []string                            `json:"before_run" yaml:"before_run"`
	AfterRun     []string                            `json:"after_run" yaml:"after_run"`
	Environments []envmanModels.EnvironmentItemModel `json:"envs" yaml:"envs"`
	Steps        []StepListItemModel                 `json:"steps" yaml:"steps"`
}

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

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
	StartTime               time.Time
	SuccessSteps            []StepRunResultsModel
	FailedSteps             []StepRunResultsModel
	FailedNotImportantSteps []StepRunResultsModel
	SkippedSteps            []StepRunResultsModel
}

// StepRunResultsModel ...
type StepRunResultsModel struct {
	StepName string
	Error    error
	ExitCode int
}
