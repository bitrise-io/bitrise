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
	EnvKey            string `yaml:"env_key"`
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

// ToBitriseConfigSerializeModel ...
func (brDataModel BitriseDataModel) ToBitriseConfigSerializeModel() BitriseConfigSerializeModel {
	workflowConfs := map[string]WorkflowSerializeModel{}
	for key, aWorkflow := range brDataModel.Workflows {
		workflowConfs[key] = aWorkflow.ToWorkflowSerializeModel()
	}

	appConf := brDataModel.App.ToAppSerializeModel()

	config := BitriseConfigSerializeModel{
		FormatVersion: brDataModel.FormatVersion,
		App:           appConf,
		Workflows:     workflowConfs,
	}

	return config
}

// ToStepSerializeModel ...
func (step StepModel) ToStepSerializeModel() StepSerializeModel {
	confInputs := []EnvironmentItemSerializeModel{}
	for _, itm := range step.Inputs {
		confInputs = append(confInputs, itm.ToEnvironmentItemSerializeModel())
	}
	confOutputs := []EnvironmentItemSerializeModel{}
	for _, itm := range step.Outputs {
		confOutputs = append(confOutputs, itm.ToEnvironmentItemSerializeModel())
	}

	return StepSerializeModel{
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

// MergeWith ...
func (envItm *EnvironmentItemModel) MergeWith(override EnvironmentItemModel) {
	envItm.Value = mergeString(envItm.Value, override.Value)
	envItm.Title = mergeString(envItm.Title, override.Title)
	envItm.Description = mergeString(envItm.Description, override.Description)
	envItm.ValueOptions = mergeStringSlice(envItm.ValueOptions, override.ValueOptions)
	envItm.IsRequired = override.IsRequired
	envItm.IsExpand = override.IsExpand
	envItm.IsDontChangeValue = override.IsDontChangeValue
}

func mergeEnvironmentItemModels(refItem, overwItem []EnvironmentItemModel) ([]EnvironmentItemModel, error) {
	for idx, aRefItm := range refItem {
		for _, aOverwItem := range overwItem {
			if aRefItm.EnvKey == aOverwItem.EnvKey {
				aRefItm.MergeWith(aOverwItem)
				refItem[idx] = aRefItm
			}
		}
	}
	return refItem, nil
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
