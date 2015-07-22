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
	Environments []stepmanModels.EnvironmentItemModel `json:"environments" yaml:"environments"`
}

// WorkflowModel ...
type WorkflowModel struct {
	Environments []stepmanModels.EnvironmentItemModel `json:"environments" yaml:"environments"`
	Steps        []StepListItemModel                  `json:"steps" yaml:"steps"`
}

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

// StepIDData ...
// structured representation of a composite-step-id
//  a composite step id is: step-lib-source::step-id@1.0.0
type StepIDData struct {
	ID            string
	Version       string
	SteplibSource string
}
