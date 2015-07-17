package models

import (
	"testing"
)

/*
// EnvironmentItemOptionsFileModel ...
type EnvironmentItemOptionsFileModel struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// EnvironmentItemFileModel ...
type EnvironmentItemFileModel map[string]interface{}

// StepSourceModel ...
type StepSourceModel struct {
	Git string `json:"git" yaml:"git"`
}

// StepFileModel ...
type StepFileModel struct {
	ID                  string                     `json:"id" yaml:"id"`
	SteplibSource       string                     `json:"steplib_source" yaml:"steplib_source"`
	VersionTag          string                     `json:"version_tag" yaml:"version_tag"`
	Name                string                     `json:"name" yaml:"name"`
	Description         string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Website             string                     `json:"website" yaml:"website"`
	ForkURL             string                     `json:"fork_url,omitempty" yaml:"fork_url,omitempty"`
	Source              StepSourceModel            `json:"source" yaml:"source"`
	HostOsTags          []string                   `json:"host_os_tags,omitempty" yaml:"host_os_tags,omitempty"`
	ProjectTypeTags     []string                   `json:"project_type_tags,omitempty" yaml:"project_type_tags,omitempty"`
	TypeTags            []string                   `json:"type_tags,omitempty" yaml:"type_tags,omitempty"`
	IsRequiresAdminUser bool                       `json:"is_requires_admin_user,omitempty" yaml:"is_requires_admin_user,omitempty"`
	Inputs              []EnvironmentItemFileModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemFileModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepListItemFile ...
type StepListItemFile map[string]StepFileModel

// WorkflowFileModel ...
type WorkflowFileModel struct {
	Environments []EnvironmentItemFileModel `json:"environments"`
	Steps        []StepListItemFile         `json:"steps"`
}

// AppFileModel ...
type AppFileModel struct {
	Environments []EnvironmentItemFileModel `json:"environments" yaml:"environments"`
}

// BitriseConfigFileModel ...
type BitriseConfigFileModel struct {
	FormatVersion string                       `json:"format_version" yaml:"format_version"`
	App           AppFileModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowFileModel `json:"workflows" yaml:"workflows"`
}
*/

func TestConvertBitriseConfig(t *testing.T) {
	defaultTrue := true
	defaultFale := false

	appEnv := EnvironmentItemFileModel{
		"TEST_KEY": "test value",
		"opts": EnvironmentItemOptionsFileModel{
			Title:        "title",
			Description:  "descr",
			ValueOptions: []string{"1tes", "w"},
			IsRequired:   &defaultTrue,
			IsExpand:     &defaultFale,
		},
	}

	stepList := StepListItemFile{
		"Step1": StepFileModel{
			ID:            "id",
			SteplibSource: "steplib",
			Source: StepSourceModel{
				Git: "https://git/url",
			},
			HostOsTags:          []string{"osx"},
			ProjectTypeTags:     []string{"ios"},
			TypeTags:            []string{"some-cat"},
			IsRequiresAdminUser: true,
			Inputs:              []EnvironmentItemFileModel{},
			Outputs:             []EnvironmentItemFileModel{},
		},
	}

	workflowFileMap := map[string]WorkflowFileModel{
		"test": WorkflowFileModel{
			Environments: []EnvironmentItemFileModel{appEnv},
			Steps:        []StepListItemFile{stepList},
		},
	}

	appFile := AppFileModel{
		Environments: []EnvironmentItemFileModel{appEnv},
	}

	configFile := BitriseConfigFileModel{
		FormatVersion: "0.0.1",
		App:           appFile,
		Workflows:     workflowFileMap,
	}

	if configFile.FormatVersion == "" {

	}
}
