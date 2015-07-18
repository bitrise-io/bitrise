package models

import (
	"testing"
)

/*
// EnvironmentItemOptionsConfigModel ...
type EnvironmentItemOptionsConfigModel struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Description       string   `json:"description,omitempty" yaml:"description,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// EnvironmentItemConfigModel ...
type EnvironmentItemConfigModel map[string]interface{}

// StepSourceModel ...
type StepSourceModel struct {
	Git string `json:"git" yaml:"git"`
}

// StepConfigModel ...
type StepConfigModel struct {
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
	Inputs              []EnvironmentItemConfigModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs             []EnvironmentItemConfigModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// StepListItemConfigModel ...
type StepListItemConfigModel map[string]StepConfigModel

// WorkflowConfigModel ...
type WorkflowConfigModel struct {
	Environments []EnvironmentItemConfigModel `json:"environments"`
	Steps        []StepListItemConfigModel         `json:"steps"`
}

// AppConfigModel ...
type AppConfigModel struct {
	Environments []EnvironmentItemConfigModel `json:"environments" yaml:"environments"`
}

// BitriseConfigModel ...
type BitriseConfigModel struct {
	FormatVersion string                       `json:"format_version" yaml:"format_version"`
	App           AppConfigModel                 `json:"app" yaml:"app"`
	Workflows     map[string]WorkflowConfigModel `json:"workflows" yaml:"workflows"`
}
*/

func TestConvertBitriseConfig(t *testing.T) {
	defaultTrue := true
	defaultFale := false

	appEnv := EnvironmentItemConfigModel{
		"TEST_KEY": "test value",
		"opts": EnvironmentItemOptionsConfigModel{
			Title:        "title",
			Description:  "descr",
			ValueOptions: []string{"1tes", "w"},
			IsRequired:   &defaultTrue,
			IsExpand:     &defaultFale,
		},
	}

	stepList := StepListItemConfigModel{
		"https://git/url::step-id@1.2.3": StepConfigModel{
			Source: StepSourceModel{
				Git: "https://git/url",
			},
			HostOsTags:          []string{"osx"},
			ProjectTypeTags:     []string{"ios"},
			TypeTags:            []string{"some-cat"},
			IsRequiresAdminUser: true,
			Inputs:              []EnvironmentItemConfigModel{},
			Outputs:             []EnvironmentItemConfigModel{},
		},
	}

	workflowFileMap := map[string]WorkflowConfigModel{
		"test": WorkflowConfigModel{
			Environments: []EnvironmentItemConfigModel{appEnv},
			Steps:        []StepListItemConfigModel{stepList},
		},
	}

	appFile := AppConfigModel{
		Environments: []EnvironmentItemConfigModel{appEnv},
	}

	configFile := BitriseConfigModel{
		FormatVersion: "0.0.1",
		App:           appFile,
		Workflows:     workflowFileMap,
	}

	if configFile.FormatVersion == "" {

	}
}
