package bitrise

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

type WorkFlowJsonStruct struct {
	FormatVersion string           `json:"format_version"`
	Environments  []string         `json:"environments"`
	Steps         []StepJsonStruct `json:"steps"`
}
