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
	ID                  string
	SteplibSource       string
	VersionTag          string
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

// BitriseConfigModel ...
type BitriseConfigModel struct {
	FormatVersion string
	App           AppModel
	Workflows     map[string]WorkflowModel
}

// -------------------
// --- Util

// MergeWith ...
func (specStep *StepModel) MergeWith(workflowStep StepModel) error {
	specStep.ID = mergeString(specStep.ID, workflowStep.ID)
	specStep.SteplibSource = mergeString(specStep.SteplibSource, workflowStep.SteplibSource)
	specStep.VersionTag = mergeString(specStep.VersionTag, workflowStep.VersionTag)
	specStep.Name = mergeString(specStep.Name, workflowStep.Name)
	specStep.Description = mergeString(specStep.Description, workflowStep.Description)
	specStep.Website = mergeString(specStep.Website, workflowStep.Website)
	specStep.ForkURL = mergeString(specStep.ForkURL, workflowStep.ForkURL)
	specStep.Source = mergeStepSourceModel(specStep.Source, workflowStep.Source)
	specStep.HostOsTags = mergeStringSlice(specStep.HostOsTags, workflowStep.HostOsTags)
	specStep.ProjectTypeTags = mergeStringSlice(specStep.ProjectTypeTags, workflowStep.ProjectTypeTags)
	specStep.TypeTags = mergeStringSlice(specStep.TypeTags, workflowStep.TypeTags)
	specStep.IsRequiresAdminUser = workflowStep.IsRequiresAdminUser

	inputs, err := mergeEnvironmentItemModels(specStep.Inputs, workflowStep.Inputs)
	if err != nil {
		return err
	}
	specStep.Inputs = inputs

	outputs, err := mergeEnvironmentItemModels(specStep.Outputs, workflowStep.Outputs)
	if err != nil {
		return err
	}
	specStep.Outputs = outputs

	return nil
}

func mergeBoolPtr(reference, override *bool) *bool {
	if override != nil {
		return override
	}
	return reference
}

func (env *EnvironmentItemModel) mergeEnvironmentItemModel(override EnvironmentItemModel) {
	*env = EnvironmentItemModel{
		Value:             mergeString(env.Value, override.Value),
		Title:             mergeString(env.Title, override.Title),
		Description:       mergeString(env.Description, override.Description),
		ValueOptions:      mergeStringSlice(env.ValueOptions, override.ValueOptions),
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
