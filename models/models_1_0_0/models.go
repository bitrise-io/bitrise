package models

import (
	"errors"

	"github.com/bitrise-io/goinp/goinp"
)

const (
	// OptionsKey ...
	OptionsKey string = "opts"
	//DefaultIsRequired ...
	DefaultIsRequired bool = false
	// DefaultIsExpand ...
	DefaultIsExpand bool = true
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue bool = false
)

// -------------------
// --- Models

// EnvironmentItemModel ...
type EnvironmentItemModel struct {
	Key               string
	Value             string
	Title             string
	Description       string
	ValueOptions      []string
	IsRequired        bool
	IsExpand          bool
	IsDontChangeValue bool
}

// StepModel ...
type StepModel struct {
	Name                string
	Description         string
	Website             string
	ForkURL             string
	Source              StepSourceModel
	HostOsTags          []string
	ProjectTypeTags     []string
	TypeTags            []string
	IsRequiresAdminUser bool
	Inputs              []EnvironmentItemModel
	Outputs             []EnvironmentItemModel
}

// StepListItem ...
type StepListItem map[string]StepModel

// WorkflowModel ...
type WorkflowModel struct {
	Environments []EnvironmentItemModel
	Steps        []StepListItem
}

// AppModel ...
type AppModel struct {
	Environments []EnvironmentItemModel
}

// BitriseDataModel ...
type BitriseDataModel struct {
	FormatVersion string
	App           AppModel
	Workflows     map[string]WorkflowModel
}

// -------------------
// --- Converters

// ToBitriseConfigModel ...
func (brDataModel BitriseDataModel) ToBitriseConfigModel() (BitriseConfigModel, error) {
	workflowConfs := map[string]WorkflowConfigModel{}
	for key, aWorkflow := range brDataModel.Workflows {
		workflowConfs[key] = aWorkflow.ToWorkflowConfigModel()
	}

	appConf := brDataModel.App.ToAppConfigModel()

	config := BitriseConfigModel{
		FormatVersion: brDataModel.FormatVersion,
		App:           appConf,
		Workflows:     workflowConfs,
	}

	return config, nil
}

// ToStepConfigModel ...
func (step StepModel) ToStepConfigModel() StepConfigModel {
	confInputs := []EnvironmentItemConfigModel{}
	for _, itm := range step.Inputs {
		confInputs = append(confInputs, itm.ToEnvironmentItemConfigModel())
	}
	confOutputs := []EnvironmentItemConfigModel{}
	for _, itm := range step.Outputs {
		confOutputs = append(confOutputs, itm.ToEnvironmentItemConfigModel())
	}

	return StepConfigModel{
		Name:                step.Name,
		Description:         step.Description,
		Website:             step.Website,
		ForkURL:             step.ForkURL,
		Source:              step.Source,
		HostOsTags:          step.HostOsTags,
		ProjectTypeTags:     step.ProjectTypeTags,
		TypeTags:            step.TypeTags,
		IsRequiresAdminUser: step.IsRequiresAdminUser,
		Inputs:              confInputs,
		Outputs:             confOutputs,
	}
}

// ToWorkflowConfigModel ...
func (wfModel WorkflowModel) ToWorkflowConfigModel() WorkflowConfigModel {
	// // WorkflowConfigModel ...
	// type WorkflowConfigModel struct {
	// 	Environments []EnvironmentItemConfigModel `json:"environments"`
	// 	Steps        []StepListItemConfigModel    `json:"steps"`
	// }
	//
	// type StepListItemConfigModel map[string]StepConfigModel

	environments := []EnvironmentItemConfigModel{}
	for _, env := range wfModel.Environments {
		environments = append(environments, env.ToEnvironmentItemConfigModel())
	}

	steps := []StepListItemConfigModel{}
	for _, stepListFile := range wfModel.Steps {
		stepList := StepListItemConfigModel{}
		for key, aStep := range stepListFile {
			stepList[key] = aStep.ToStepConfigModel()
		}
		steps = append(steps, stepList)
	}

	worflow := WorkflowConfigModel{
		Environments: environments,
		Steps:        steps,
	}

	return worflow
}

// ToEnvironmentItemConfigModel ...
func (envItm EnvironmentItemModel) ToEnvironmentItemConfigModel() EnvironmentItemConfigModel {
	return EnvironmentItemConfigModel{
		envItm.Key: envItm.Value,
		OptionsKey: EnvironmentItemOptionsConfigModel{
			Title:             envItm.Title,
			Description:       envItm.Description,
			ValueOptions:      envItm.ValueOptions,
			IsRequired:        &envItm.IsRequired,
			IsExpand:          &envItm.IsExpand,
			IsDontChangeValue: &envItm.IsDontChangeValue,
		},
	}
}

// ToAppConfigModel ...
func (appData AppModel) ToAppConfigModel() AppConfigModel {
	environments := []EnvironmentItemConfigModel{}
	for _, envItm := range appData.Environments {
		environments = append(environments, envItm.ToEnvironmentItemConfigModel())
	}

	app := AppConfigModel{
		Environments: environments,
	}

	return app
}

// -------------------
// --- Util

// MergeWith ...
func (step *StepModel) MergeWith(workflowStep StepModel) error {
	step.Name = mergeString(step.Name, workflowStep.Name)
	step.Description = mergeString(step.Description, workflowStep.Description)
	step.Website = mergeString(step.Website, workflowStep.Website)
	step.ForkURL = mergeString(step.ForkURL, workflowStep.ForkURL)
	step.Source = mergeStepSourceModel(step.Source, workflowStep.Source)
	step.HostOsTags = mergeStringSlice(step.HostOsTags, workflowStep.HostOsTags)
	step.ProjectTypeTags = mergeStringSlice(step.ProjectTypeTags, workflowStep.ProjectTypeTags)
	step.TypeTags = mergeStringSlice(step.TypeTags, workflowStep.TypeTags)
	step.IsRequiresAdminUser = workflowStep.IsRequiresAdminUser

	inputs, err := mergeEnvironmentItemModels(step.Inputs, workflowStep.Inputs)
	if err != nil {
		return err
	}
	step.Inputs = inputs

	outputs, err := mergeEnvironmentItemModels(step.Outputs, workflowStep.Outputs)
	if err != nil {
		return err
	}
	step.Outputs = outputs

	return nil
}

func mergeBoolPtr(reference, override *bool) *bool {
	if override != nil {
		return override
	}
	return reference
}

func (envItm *EnvironmentItemModel) mergeEnvironmentItemModel(override EnvironmentItemModel) {
	*envItm = EnvironmentItemModel{
		Value:             mergeString(envItm.Value, override.Value),
		Title:             mergeString(envItm.Title, override.Title),
		Description:       mergeString(envItm.Description, override.Description),
		ValueOptions:      mergeStringSlice(envItm.ValueOptions, override.ValueOptions),
		IsRequired:        override.IsRequired,
		IsExpand:          override.IsExpand,
		IsDontChangeValue: override.IsDontChangeValue,
	}
}

func mergeEnvironmentItemModels(reference, override []EnvironmentItemModel) ([]EnvironmentItemModel, error) {
	for idx, referenceEnv := range reference {
		for _, overrideEnv := range override {
			if referenceEnv.Key == overrideEnv.Key {
				referenceEnv.mergeEnvironmentItemModel(overrideEnv)
				reference[idx] = referenceEnv
			}
		}
	}
	return reference, nil
}

func mergeStringSlice(reference, override []string) []string {
	if len(override) > 0 {
		return override
	}
	return reference
}

func mergeStepSourceModel(reference, override StepSourceModel) StepSourceModel {
	if override.Git != "" {
		return override
	}
	return reference
}

func mergeString(reference, override string) string {
	if override != "" {
		return override
	}
	return reference
}

func parseBoolWithDefault(stringOrBool interface{}, defaultValue bool) (bool, error) {
	if stringOrBool == nil {
		return defaultValue, nil
	}

	boolValue := defaultValue
	var err error
	var ok bool
	switch stringOrBool.(type) {
	case string:
		if stringValue, ok := stringOrBool.(string); ok == false {
			return defaultValue, errors.New("Failed to cast to string")
		} else if boolValue, err = goinp.ParseBool(stringValue); err != nil {
			return defaultValue, errors.New("Failed to parse bool")
		}
	case bool:
		if boolValue, ok = stringOrBool.(bool); ok == false {
			return defaultValue, errors.New("Failed to cast to bool")
		}
	default:
		return defaultValue, errors.New("Failed to parse: Unknown type")
	}
	return boolValue, nil
}

func defaultEnvironmentItemModel() EnvironmentItemModel {
	env := EnvironmentItemModel{
		Key:               "",
		Value:             "",
		Title:             "",
		Description:       "",
		ValueOptions:      []string{},
		IsRequired:        DefaultIsRequired,
		IsExpand:          DefaultIsExpand,
		IsDontChangeValue: DefaultIsDontChangeValue,
	}
	return env
}
