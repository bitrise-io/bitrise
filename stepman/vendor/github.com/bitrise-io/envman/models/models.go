package models

// EnvironmentItemOptionsModel ...
type EnvironmentItemOptionsModel struct {
	// These fields are processed by envman at envman run
	IsExpand    *bool `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	SkipIfEmpty *bool `json:"skip_if_empty,omitempty" yaml:"skip_if_empty,omitempty"`
	// These fields used only by bitrise
	Title             *string  `json:"title,omitempty" yaml:"title,omitempty"`
	Description       *string  `json:"description,omitempty" yaml:"description,omitempty"`
	Summary           *string  `json:"summary,omitempty" yaml:"summary,omitempty"`
	Category          *string  `json:"category,omitempty" yaml:"category,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
	IsTemplate        *bool    `json:"is_template,omitempty" yaml:"is_template,omitempty"`
}

// EnvironmentItemModel ...
type EnvironmentItemModel map[string]interface{}

// EnvsSerializeModel ...
type EnvsSerializeModel struct {
	Envs []EnvironmentItemModel `json:"envs",yaml:"envs"`
}

// EnvsJSONListModel ...
type EnvsJSONListModel map[string]string
