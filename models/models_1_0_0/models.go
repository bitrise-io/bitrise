package models

import (
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// -------------------
// --- Bitrise-cli models

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion        string                   `json:"format_version" yaml:"format_version"`
	DefaultStepLibSource string                   `json:"default_step_lib_source" yaml:"default_step_lib_source"`
	App                  AppModel                 `json:"app" yaml:"app"`
	Workflows            map[string]WorkflowModel `json:"workflows" yaml:"workflows"`
}

// AppModel ...
type AppModel struct {
	Environments []stepmanModels.EnvironmentItemModel `json:"envs" yaml:"envs"`
}

// WorkflowModel ...
type WorkflowModel struct {
	BeforeWorkflow string                               `json:"before_run" yaml:"before_run"`
	AfterWorkflow  string                               `json:"after_run" yaml:"after_run"`
	Environments   []stepmanModels.EnvironmentItemModel `json:"envs" yaml:"envs"`
	Steps          []StepListItemModel                  `json:"steps" yaml:"steps"`
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

// StepRunResultsModel ...
type StepRunResultsModel struct {
	TotalStepCount          int
	FailedSteps             []FailedStepModel
	FailedNotImportantSteps []FailedStepModel
	SkippedSteps            []FailedStepModel
}

// FailedStepModel ...
type FailedStepModel struct {
	StepName string
	Error    error
}
