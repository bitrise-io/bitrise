package main

import ()

// Json
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

// YML
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

type StepCollectionYmlStruct struct {
	FormatVersion        string          `yaml:"format_version"`
	GeneratedAtTimeStamp string          `yaml:"generated_at_timestamp"`
	Steps                []StepYmlStruct `yaml:"steps"`
	SteplibSource        string          `yaml:"steplib_source"`
}
