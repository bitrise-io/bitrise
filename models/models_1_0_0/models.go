package models

import (
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// -------------------
// --- Bitrise-cli models

// StepListItemModel ...
type StepListItemModel map[string]stepmanModels.StepModel

// AppModel ...
type AppModel struct {
	Environments []stepmanModels.EnvironmentItemModel `json:"environments" yaml:"environments"`
}

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion string                   `json:"format_version" yaml:"format_version"`
	App           AppModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowModel `json:"workflows" yaml:"workflows"`
}

// WorkflowModel ...
type WorkflowModel struct {
	FormatVersion string                               `json:"format_version"`
	Environments  []stepmanModels.EnvironmentItemModel `json:"environments"`
	Steps         []StepListItemModel                  `json:"steps"`
}
