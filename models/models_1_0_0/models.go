package models

import (
	"errors"

	"github.com/bitrise-io/goinp/goinp"
)

// -------------------
// --- YML Models

// EnvironmentYMLItemModel ...
type EnvironmentYMLItemModel map[string]interface{}

// StepYMLModel ...
type StepYMLModel struct {
	ID                  string                    `json:"id" yaml:"id"`
	SteplibSource       string                    `json:"steplib_source" yaml:"steplib_source"`
	VersionTag          string                    `json:"version_tag" yaml:"version_tag"`
	Name                string                    `json:"name" yaml:"name"`
	Description         string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string                    `json:"website" yaml:"website"`
	ForkURL             string                    `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel           `json:"source" yaml:"source"`
	HostOsTags          []string                  `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string                  `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string                  `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool                      `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []EnvironmentYMLItemModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentYMLItemModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepListYMLItem ...
type StepListYMLItem map[string]StepYMLModel

// WorkflowYMLModel ...
type WorkflowYMLModel struct {
	Environments []EnvironmentYMLItemModel `json:"environments"`
	Steps        []StepListYMLItem         `json:"steps"`
}

// AppYMLModel ...
type AppYMLModel struct {
	Environments []EnvironmentYMLItemModel `json:"environments" yaml:"environments"`
}

// BitriseConfigYMLModel ...
type BitriseConfigYMLModel struct {
	FormatVersion string                      `json:"format_version" yaml:"format_version"`
	App           AppYMLModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowYMLModel `json:"workflows" yaml:"workflows"`
}

// -------------------
// --- Models

// InputModel ...
type InputModel struct {
	MappedTo          string   `json:"mapped_to,omitempty" yaml:"mapped_to,omitempty"`
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	Value             string   `json:"value,omitempty" yaml:"value,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// StepSourceModel ...
type StepSourceModel struct {
	Git string `json:"git" yaml:"git"`
}

// StepModel ...
type StepModel struct {
	ID                  string          `json:"id" yaml:"id"`
	SteplibSource       string          `json:"steplib_source" yaml:"steplib_source"`
	VersionTag          string          `json:"version_tag" yaml:"version_tag"`
	Name                string          `json:"name" yaml:"name"`
	Description         string          `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string          `json:"website" yaml:"website"`
	ForkURL             string          `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel `json:"source" yaml:"source"`
	HostOsTags          []string        `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string        `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string        `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool            `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []InputModel    `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []InputModel    `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepListItem ...
type StepListItem map[string]StepModel

// WorkflowModel ...
type WorkflowModel struct {
	Environments []InputModel   `json:"environments"`
	Steps        []StepListItem `json:"steps"`
}

// AppModel ...
type AppModel struct {
	Environments []InputModel `json:"environments" yaml:"environments"`
}

// BitriseConfigModel ...
type BitriseConfigModel struct {
	FormatVersion string                   `json:"format_version" yaml:"format_version"`
	App           AppModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowModel `json:"workflows" yaml:"workflows"`
}

// -------------------
// --- Struct methods

// GetStepIDStepDataPair ...
func (stepListItm StepListItem) GetStepIDStepDataPair() (string, StepModel, error) {
	if len(stepListItm) > 1 {
		return "", StepModel{}, errors.New("StepListItem contains more than 1 key-value pair!")
	}
	for key, value := range stepListItm {
		return key, value, nil
	}
	return "", StepModel{}, errors.New("StepListItem does not contain a key-value pair!")
}

// ToInputModel ...
func (environmentYMLItem EnvironmentYMLItemModel) ToInputModel() (InputModel, error) {
	inputModel := defaultInputModel()
	for key, value := range environmentYMLItem {
		var ok bool
		switch key {
		case "title":
			inputModel.Title, ok = value.(string)
			if ok == false {
				return InputModel{}, errors.New("Failed to cast title")
			}
		case "description":
			inputModel.Description, ok = value.(string)
			if ok == false {
				return InputModel{}, errors.New("Failed to cast description")
			}
		case "value_options":
			inputModel.ValueOptions, ok = value.([]string)
			if ok == false {
				return InputModel{}, errors.New("Failed to cast value_options")
			}
		case "is_required":
			boolValue, err := parseBoolWithDefault(value, false)
			if err != nil {
				return InputModel{}, err
			}
			inputModel.IsRequired = &boolValue
		case "is_expand":
			boolValue, err := parseBoolWithDefault(value, true)
			if err != nil {
				return InputModel{}, err
			}
			inputModel.IsExpand = &boolValue
		case "is_dont_change_value":
			boolValue, err := parseBoolWithDefault(value, false)
			if err != nil {
				return InputModel{}, err
			}
			inputModel.IsDontChangeValue = &boolValue
		default:
			inputModel.MappedTo = key
			inputModel.Value, ok = value.(string)
			if ok == false {
				return InputModel{}, errors.New("Failed to cast value")
			}
		}
	}
	return inputModel, nil
}

// ToStepModel ...
func (stepYML StepYMLModel) ToStepModel() (StepModel, error) {
	inputs := []InputModel{}
	for _, envYMLItem := range stepYML.Inputs {
		input, err := envYMLItem.ToInputModel()
		if err != nil {
			return StepModel{}, err
		}
		inputs = append(inputs, input)
	}

	outputs := []InputModel{}
	for _, envYMLItem := range stepYML.Outputs {
		output, err := envYMLItem.ToInputModel()
		if err != nil {
			return StepModel{}, err
		}
		outputs = append(outputs, output)
	}

	step := StepModel{
		ID:                  stepYML.ID,
		SteplibSource:       stepYML.SteplibSource,
		VersionTag:          stepYML.VersionTag,
		Name:                stepYML.Name,
		Description:         stepYML.Description,
		Website:             stepYML.Website,
		ForkURL:             stepYML.ForkURL,
		Source:              stepYML.Source,
		HostOsTags:          stepYML.HostOsTags,
		ProjectTypeTags:     stepYML.ProjectTypeTags,
		TypeTags:            stepYML.TypeTags,
		IsRequiresAdminUser: stepYML.IsRequiresAdminUser,
		Inputs:              inputs,
		Outputs:             outputs,
	}

	return step, nil
}

// ToStepListItem ...
func (stepListYMLItem StepListYMLItem) ToStepListItem() (StepListItem, error) {
	stepListItem := StepListItem{}
	for key, value := range stepListYMLItem {
		stepModel, err := value.ToStepModel()
		if err != nil {
			return StepListItem{}, err
		}
		stepListItem[key] = stepModel
	}
	return stepListItem, nil
}

// ToWorkflowModel ...
func (workflowYMLModel WorkflowYMLModel) ToWorkflowModel() (WorkflowModel, error) {
	environments := []InputModel{}
	for _, envYML := range workflowYMLModel.Environments {
		input, err := envYML.ToInputModel()
		if err != nil {
			return WorkflowModel{}, err
		}
		environments = append(environments, input)
	}

	steps := []StepListItem{}
	for _, stepListItemYML := range workflowYMLModel.Steps {
		stepListItem, err := stepListItemYML.ToStepListItem()
		if err != nil {
			return WorkflowModel{}, err
		}
		steps = append(steps, stepListItem)
	}

	workflow := WorkflowModel{
		Environments: environments,
		Steps:        steps,
	}

	return workflow, nil
}

// ToAppModel ...
func (appYml AppYMLModel) ToAppModel() (AppModel, error) {
	environments := []InputModel{}
	for _, envYML := range appYml.Environments {
		input, err := envYML.ToInputModel()
		if err != nil {
			return AppModel{}, err
		}
		environments = append(environments, input)
	}

	appModel := AppModel{
		Environments: environments,
	}
	return appModel, nil
}

// ToBitriseConfigModel ...
func (bitriseConfigYML BitriseConfigYMLModel) ToBitriseConfigModel() (BitriseConfigModel, error) {
	workflows := map[string]WorkflowModel{}
	for key, value := range bitriseConfigYML.Workflows {
		workflow, err := value.ToWorkflowModel()
		if err != nil {
			return BitriseConfigModel{}, err
		}
		workflows[key] = workflow
	}

	app, err := bitriseConfigYML.App.ToAppModel()
	if err != nil {
		return BitriseConfigModel{}, err
	}

	config := BitriseConfigModel{
		FormatVersion: bitriseConfigYML.FormatVersion,
		App:           app,
		Workflows:     workflows,
	}

	return config, nil
}

// -------------------
// --- Util

// MergeWith ...
func (specStep *StepModel) MergeWith(workflowStep StepModel) {
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
	specStep.Inputs = mergeInputModels(specStep.Inputs, workflowStep.Inputs)
	specStep.Outputs = mergeInputModels(specStep.Outputs, workflowStep.Outputs)
}

func mergeBoolPtr(reference, override *bool) *bool {
	if override != nil {
		return override
	}
	return reference
}

func (reference *InputModel) mergeInputModel(override InputModel) {
	reference.MappedTo = mergeString(reference.MappedTo, override.MappedTo)
	reference.Title = mergeString(reference.Title, override.Title)
	reference.Description = mergeString(reference.Description, override.Description)
	reference.Value = mergeString(reference.Value, override.Value)
	reference.ValueOptions = mergeStringSlice(reference.ValueOptions, override.ValueOptions)
	reference.IsRequired = mergeBoolPtr(reference.IsRequired, override.IsRequired)
	reference.IsExpand = mergeBoolPtr(reference.IsExpand, override.IsExpand)
	reference.IsDontChangeValue = mergeBoolPtr(reference.IsDontChangeValue, override.IsDontChangeValue)
}

func mergeInputModels(reference, override []InputModel) []InputModel {
	for idx, referenceInput := range reference {
		for _, overrideInput := range override {
			if referenceInput.MappedTo == overrideInput.MappedTo {
				referenceInput.mergeInputModel(overrideInput)
				reference[idx] = referenceInput
			}
		}
	}
	return reference
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

func defaultInputModel() InputModel {
	defaultString := ""
	defaultFalse := false
	defaultTrue := true
	inputModel := InputModel{
		MappedTo:          defaultString,
		Title:             defaultString,
		Description:       defaultString,
		Value:             defaultString,
		ValueOptions:      []string{},
		IsRequired:        &defaultFalse,
		IsExpand:          &defaultTrue,
		IsDontChangeValue: &defaultFalse,
	}
	return inputModel
}
