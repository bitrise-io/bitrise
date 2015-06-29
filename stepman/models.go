package main

import ()

// Json
type InputJsonStruct struct {
	MappedTo          string   `json:"mapped_to"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Value             string   `json:"value"`
	ValueOptions      []string `json:"value_options"`
	IsRequired        bool     `json:"is_required"`
	IsExpand          bool     `json:"is_expand"`
	IsDontChangeValue bool     `json:"is_dont_change_value"`
}

type OutputJsonStruct struct {
	MappedTo    string `json:"mapped_to"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type StepJsonStruct struct {
	Id                  string             `json:"id"`
	StepLibSource       string             `json:"steplib_source"`
	VersionTag          string             `json:"version_tag"`
	Name                string             `json:"name"`
	Description         string             `json:"description"`
	Website             string             `json:"website"`
	ForkUrl             string             `json:"fork_url"`
	Source              map[string]string  `json:"source"`
	HostOsTags          []string           `json:"host_os_tags"`
	ProjectTypeTags     []string           `json:"project_type_tags"`
	TypeTags            []string           `json:"type_tags"`
	IsRequiresAdminUser bool               `json:"is_requires_admin_user"`
	Inputs              []InputJsonStruct  `json:"inputs"`
	Outputs             []OutputJsonStruct `json:"outputs"`
}

type StepGroupJsonStruct struct {
	Id       string           `json:"id"`
	Versions []StepJsonStruct `json:"versions"`
	Latest   StepJsonStruct   `json:"latest"`
}

type StepJsonHash map[string]StepGroupJsonStruct

type StepCollectionJsonStruct struct {
	FormatVersion        string       `json:"format_version"`
	GeneratedAtTimeStamp string       `json:"generated_at_timestamp"`
	Steps                StepJsonHash `json:"steps"`
	SteplibSource        string       `json:"steplib_source"`
}

// YML
type InputYmlStruct struct {
	MappedTo          string   `yml:"mapped_to"`
	Title             string   `yml:"title"`
	Description       string   `yml:"description"`
	Value             string   `yml:"value"`
	ValueOptions      []string `yml:"value_options"`
	IsRequired        bool     `yml:"is_required"`
	IsExpand          bool     `yml:"is_expand"`
	IsDontChangeValue bool     `yml:"is_dont_change_value"`
}

type OutputYmlStruct struct {
	MappedTo    string `yml:"mapped_to"`
	Title       string `yml:"title"`
	Description string `yml:"description"`
}

type StepYmlStruct struct {
	Name                string            `yml:"name"`
	Description         string            `yml:"description"`
	Website             string            `yml:"website"`
	ForkUrl             string            `yml:"fork_url"`
	Source              map[string]string `yml:"source"`
	HostOsTags          []string          `yml:"host_os_tags"`
	ProjectTypeTags     []string          `yml:"project_type_tags"`
	TypeTags            []string          `yml:"type_tags"`
	IsRequiresAdminUser bool              `yml:"is_requires_admin_user"`
	Inputs              []InputYmlStruct  `yml:"inputs"`
	Outputs             []OutputYmlStruct `yml:"outputs"`
}

type StepCollectionYmlStruct struct {
	FormatVersion        string          `yml:"format_version"`
	GeneratedAtTimeStamp string          `yml:"generated_at_timestamp"`
	Steps                []StepYmlStruct `yml:"steps"`
	SteplibSource        string          `yml:"steplib_source"`
}
