package models

// EnvironmentItemOptionsModel ...
type EnvironmentItemOptionsModel struct {
	Title             *string  `json:"title,omitempty" yaml:"title,omitempty"`
	Description       *string  `json:"description,omitempty" yaml:"description,omitempty"`
	Summary           *string  `json:"summary,omitempty" yaml:"summary,omitempty"`
	ValueOptions      []string `json:"value_options,omitempty" yaml:"value_options,omitempty"`
	IsRequired        *bool    `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	IsExpand          *bool    `json:"is_expand,omitempty" yaml:"is_expand,omitempty"`
	IsDontChangeValue *bool    `json:"is_dont_change_value,omitempty" yaml:"is_dont_change_value,omitempty"`
}

// EnvironmentItemModel ...
type EnvironmentItemModel map[string]interface{}

// EnvsYMLModel ...
type EnvsYMLModel struct {
	Envs []EnvironmentItemModel `yaml:"envs"`
}
