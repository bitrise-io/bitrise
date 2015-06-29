package main

import ()

//
// Models
//

// - Json
type InputJsonStruct struct {
	MappedTo          *string   `json:"mapped_to,omitempty"`
	Title             *string   `json:"title,omitempty"`
	Description       *string   `json:"description,omitempty"`
	Value             *string   `json:"value,omitempty"`
	ValueOptions      *[]string `json:"value_options,omitempty"`
	IsRequired        *bool     `json:"is_required,omitempty"`
	IsExpand          *bool     `json:"is_expand,omitempty"`
	IsDontChangeValue *bool     `json:"is_dont_change_value,omitempty"`
}

type OutputJsonStruct struct {
	MappedTo    *string `json:"mapped_to,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type StepJsonStruct struct {
	Id                  string              `json:"id"`
	StepLibSource       string              `json:"steplib_source"`
	VersionTag          string              `json:"version_tag"`
	Name                string              `json:"name"`
	Description         *string             `json:"description,omitempty"`
	Website             string              `json:"website"`
	ForkUrl             *string             `json:"fork_url,omitempty"`
	Source              map[string]string   `json:"source"`
	HostOsTags          *[]string           `json:"host_os_tags,omitempty"`
	ProjectTypeTags     *[]string           `json:"project_type_tags,omitempty"`
	TypeTags            *[]string           `json:"type_tags,omitempty"`
	IsRequiresAdminUser *bool               `json:"is_requires_admin_user,omitempty"`
	Inputs              []*InputJsonStruct  `json:"inputs,omitempty"`
	Outputs             []*OutputJsonStruct `json:"outputs,omitempty"`
}

type StepGroupJsonStruct struct {
	Id       string           `json:"id"`
	Versions []StepJsonStruct `json:"versions"`
	Latest   StepJsonStruct   `json:"latest"`
}

type StepJsonHash map[string]StepGroupJsonStruct

type StepCollectionJsonStruct struct {
	FormatVersion        string       `json:"format_version"`
	GeneratedAtTimeStamp int64        `json:"generated_at_timestamp"`
	Steps                StepJsonHash `json:"steps"`
	SteplibSource        string       `json:"steplib_source"`
}

func (stepCollection StepCollectionJsonStruct) GetStep(id, version string) (bool, StepJsonStruct) {
	versions := stepCollection.Steps[id].Versions
	if len(versions) > 0 {
		for _, step := range versions {
			if step.VersionTag == version {
				return true, step
			}
		}
	}
	return false, StepJsonStruct{}
}

// - YML
type InputYmlStruct struct {
	MappedTo          *string   `yaml:"mapped_to,omitempty"`
	Title             *string   `yaml:"title,omitempty"`
	Description       *string   `yaml:"description,omitempty"`
	Value             *string   `yaml:"value,omitempty"`
	ValueOptions      *[]string `yaml:"value_options,omitempty"`
	IsRequired        *bool     `yaml:"is_required,omitempty"`
	IsExpand          *bool     `yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool     `yaml:"is_dont_change_value,omitempty"`
}

type OutputYmlStruct struct {
	MappedTo    *string `yaml:"mapped_to,omitempty"`
	Title       *string `yaml:"title,omitempty"`
	Description *string `yaml:"description,omitempty"`
}

type StepYmlStruct struct {
	Name                string             `yaml:"name"`
	Description         *string            `yaml:"description,omitempty"`
	Website             string             `yaml:"website"`
	ForkUrl             *string            `yaml:"fork_url,omitempty"`
	Source              map[string]string  `yaml:"source"`
	HostOsTags          *[]string          `yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     *[]string          `yaml:"project_type_tags,omitempty"`
	TypeTags            *[]string          `yaml:"type_tags,omitempty"`
	IsRequiresAdminUser *bool              `yaml:"is_requires_admin_user,omitempty"`
	Inputs              []*InputYmlStruct  `yaml:"inputs,omitempty"`
	Outputs             []*OutputYmlStruct `yaml:"outputs,omitempty"`
}

//
// Converters
//
func convertToInputJsonStruct(inputYml InputYmlStruct) InputJsonStruct {
	inputJson := InputJsonStruct{
		MappedTo:          inputYml.MappedTo,
		Title:             inputYml.Title,
		Description:       inputYml.Description,
		Value:             inputYml.Value,
		ValueOptions:      inputYml.ValueOptions,
		IsRequired:        inputYml.IsRequired,
		IsExpand:          inputYml.IsExpand,
		IsDontChangeValue: inputYml.IsDontChangeValue,
	}
	return inputJson
}

func convertToInputJsonSructSlice(inputYmlSlice []*InputYmlStruct) []*InputJsonStruct {
	if len(inputYmlSlice) > 0 {
		inputJsonSlice := make([]*InputJsonStruct, len(inputYmlSlice))
		for i, inputYml := range inputYmlSlice {
			inputJson := convertToInputJsonStruct(*inputYml)
			inputJsonSlice[i] = &inputJson
		}
		return inputJsonSlice
	}
	return []*InputJsonStruct{}
}

func convertToOutputJsonStruct(outputYml OutputYmlStruct) OutputJsonStruct {
	outputJson := OutputJsonStruct{
		MappedTo:    outputYml.MappedTo,
		Title:       outputYml.Title,
		Description: outputYml.Description,
	}
	return outputJson
}

func convertToOutputJsonSructSlice(outputYmlSlice []*OutputYmlStruct) []*OutputJsonStruct {
	if len(outputYmlSlice) > 0 {
		outputJsonSlice := make([]*OutputJsonStruct, len(outputYmlSlice))
		for i, outputYml := range outputYmlSlice {
			outputJson := convertToOutputJsonStruct(*outputYml)
			outputJsonSlice[i] = &outputJson
		}
		return outputJsonSlice
	}
	return []*OutputJsonStruct{}
}

func convertToStepJsonStruct(stepYml StepYmlStruct) StepJsonStruct {
	inputsJson := convertToInputJsonSructSlice(stepYml.Inputs)

	outputsJson := convertToOutputJsonSructSlice(stepYml.Outputs)

	stepJson := StepJsonStruct{
		Name:                stepYml.Name,
		Description:         stepYml.Description,
		Website:             stepYml.Website,
		ForkUrl:             stepYml.ForkUrl,
		Source:              stepYml.Source,
		HostOsTags:          stepYml.HostOsTags,
		ProjectTypeTags:     stepYml.ProjectTypeTags,
		TypeTags:            stepYml.TypeTags,
		IsRequiresAdminUser: stepYml.IsRequiresAdminUser,
		Inputs:              inputsJson,
		Outputs:             outputsJson,
	}

	return stepJson
}
