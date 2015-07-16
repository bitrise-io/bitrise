package models

import (
	"errors"

	log "github.com/Sirupsen/logrus"
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
	Source              map[string]string         `json:"source" yaml:"source"`
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

// StepModel ...
type StepModel struct {
	ID                  string            `json:"id" yaml:"id"`
	SteplibSource       string            `json:"steplib_source" yaml:"steplib_source"`
	VersionTag          string            `json:"version_tag" yaml:"version_tag"`
	Name                string            `json:"name" yaml:"name"`
	Description         string            `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string            `json:"website" yaml:"website"`
	ForkURL             string            `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              map[string]string `json:"source" yaml:"source"`
	HostOsTags          []string          `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string          `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string          `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool              `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []InputModel      `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []InputModel      `json:"outputs,omitempty" yaml:"outputs,omitempty"`
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

// InputModel ...
func (environmentYMLItem EnvironmentYMLItemModel) InputModel() InputModel {
	inputModel := defaultInputModel()
	for key, value := range environmentYMLItem {
		if value == nil {
			continue
		}

		var ok bool
		switch key {
		case "title":
			inputModel.Title, ok = value.(string)
			if ok == false {
				log.Fatal("Failed to cast")
			}
		case "description":
			inputModel.Description, ok = value.(string)
			if ok == false {
				log.Fatal("Failed to cast")
			}
		case "value_options":
			inputModel.ValueOptions, ok = value.([]string)
			if ok == false {
				log.Fatal("Failed to cast")
			}
		case "is_required":
			boolValue := parseBool(value, false)
			inputModel.IsRequired = &boolValue
		case "is_expand":
			boolValue := parseBool(value, true)
			inputModel.IsExpand = &boolValue
		case "is_dont_change_value":
			boolValue := parseBool(value, false)
			inputModel.IsDontChangeValue = &boolValue
		default:
			inputModel.MappedTo = key
			inputModel.Value, ok = value.(string)
			if ok == false {
				log.Fatal("Failed to cast")
			}
		}
	}
	return inputModel
}

// StepModel ...
func (stepYML StepYMLModel) StepModel() StepModel {
	inputs := []InputModel{}
	for _, envYMLItem := range stepYML.Inputs {
		input := envYMLItem.InputModel()
		inputs = append(inputs, input)
	}

	outputs := []InputModel{}
	for _, envYMLItem := range stepYML.Outputs {
		output := envYMLItem.InputModel()
		outputs = append(outputs, output)
	}

	step := StepModel{}
	step.ID = stepYML.ID
	step.SteplibSource = stepYML.SteplibSource
	step.VersionTag = stepYML.VersionTag
	step.Name = stepYML.Name
	step.Description = stepYML.Description
	step.Website = stepYML.Website
	step.ForkURL = stepYML.ForkURL
	step.Source = stepYML.Source
	step.HostOsTags = stepYML.HostOsTags
	step.ProjectTypeTags = stepYML.ProjectTypeTags
	step.TypeTags = stepYML.TypeTags
	step.IsRequiresAdminUser = stepYML.IsRequiresAdminUser
	step.Inputs = inputs
	step.Outputs = outputs

	return step
}

// StepListItem ...
func (stepListYMLItem StepListYMLItem) StepListItem() StepListItem {
	stepListItem := StepListItem{}
	for key, value := range stepListYMLItem {
		stepModel := value.StepModel()
		stepListItem[key] = stepModel
	}
	return stepListItem
}

// WorkflowModel ...
func (workflowYMLModel WorkflowYMLModel) WorkflowModel() WorkflowModel {
	environments := []InputModel{}
	for _, envYML := range workflowYMLModel.Environments {
		input := envYML.InputModel()
		environments = append(environments, input)
	}

	steps := []StepListItem{}
	for _, stepListItemYML := range workflowYMLModel.Steps {
		stepListItem := stepListItemYML.StepListItem()
		steps = append(steps, stepListItem)
	}

	workflow := WorkflowModel{}
	workflow.Environments = environments
	workflow.Steps = steps

	return workflow
}

// AppModel ...
func (appYml AppYMLModel) AppModel() AppModel {
	environments := []InputModel{}
	for _, envYML := range appYml.Environments {
		input := envYML.InputModel()
		environments = append(environments, input)
	}

	appModel := AppModel{}
	appModel.Environments = environments
	return appModel
}

// BitriseConfigModel ...
func (bitriseConfigYML BitriseConfigYMLModel) BitriseConfigModel() BitriseConfigModel {
	workflows := map[string]WorkflowModel{}
	for key, value := range bitriseConfigYML.Workflows {
		workflow := value.WorkflowModel()
		workflows[key] = workflow
	}

	config := BitriseConfigModel{}
	config.FormatVersion = bitriseConfigYML.FormatVersion
	config.App = bitriseConfigYML.App.AppModel()
	config.Workflows = workflows
	return config
}

// -------------------
// --- Util

func parseBool(stringOrBool interface{}, defaultValue bool) bool {
	boolValue := defaultValue
	var err error
	var ok bool
	switch t := stringOrBool.(type) {
	case string:
		stringValue, ok := stringOrBool.(string)
		if ok == false {
			log.Fatal("Failed to cast")
		}
		if boolValue, err = goinp.ParseBool(stringValue); err != nil {
			log.Fatal("Failed to parse bool:", err)
		}
	case bool:
		boolValue, ok = stringOrBool.(bool)
		if ok == false {
			log.Fatal("Failed to cast")
		}
	default:
		log.Fatal("Failed to parse bool, type:", t)
	}
	return boolValue
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
