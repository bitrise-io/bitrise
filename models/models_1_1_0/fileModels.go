package models

import "errors"

// -------------------
// --- File models

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

// -------------------
// --- Struct methods

// GetKeyValuePair ...
func (envFile EnvironmentItemFileModel) GetKeyValuePair() (string, string, error) {
	if len(envFile) < 3 {
		for key, value := range envFile {
			if key != OptionsKey {
				valueStr, ok := value.(string)
				if ok == false {
					return "", "", errors.New("Invalid value")
				}
				return key, valueStr, nil
			}
		}
	}
	return "", "", errors.New("Invalid envFile")
}

// GetOptions ...
func (envFile EnvironmentItemFileModel) GetOptions() (EnvironmentItemOptionsFileModel, error) {
	value, found := envFile[OptionsKey]
	if !found {
		return EnvironmentItemOptionsFileModel{}, nil
	}

	options, ok := value.(EnvironmentItemOptionsFileModel)
	if !ok {
		return EnvironmentItemOptionsFileModel{}, errors.New("Invalid options")
	}

	return options, nil
}

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

// ToEnvironmentItemModel ...
func (envFile EnvironmentItemFileModel) ToEnvironmentItemModel() (EnvironmentItemModel, error) {
	key, value, err := envFile.GetKeyValuePair()
	if err != nil {
		return EnvironmentItemModel{}, err
	}

	options, err := envFile.GetOptions()
	if err != nil {
		return EnvironmentItemModel{}, err
	}

	isRequired := DefaultIsRequired
	if options.IsRequired != nil {
		isRequired = *options.IsRequired
	}

	isExpand := DefaultIsExpand
	if options.IsExpand != nil {
		isExpand = *options.IsExpand
	}

	isDontChnageValue := DefaultIsDontChangeValue
	if options.IsDontChangeValue != nil {
		isDontChnageValue = *options.IsDontChangeValue
	}

	env := EnvironmentItemModel{
		Key:               key,
		Value:             value,
		Title:             options.Title,
		Description:       options.Description,
		ValueOptions:      options.ValueOptions,
		IsRequired:        isRequired,
		IsExpand:          isExpand,
		IsDontChangeValue: isDontChnageValue,
	}

	return env, nil
}

// ToStepModel ...
func (stepFile StepFileModel) ToStepModel() (StepModel, error) {
	inputs := []EnvironmentItemModel{}
	for _, envFile := range stepFile.Inputs {
		env, err := envFile.ToEnvironmentItemModel()
		if err != nil {
			return StepModel{}, err
		}
		inputs = append(inputs, env)
	}

	outputs := []EnvironmentItemModel{}
	for _, envFile := range stepFile.Outputs {
		env, err := envFile.ToEnvironmentItemModel()
		if err != nil {
			return StepModel{}, err
		}
		outputs = append(outputs, env)
	}

	step := StepModel{
		ID:                  stepFile.ID,
		SteplibSource:       stepFile.SteplibSource,
		VersionTag:          stepFile.VersionTag,
		Name:                stepFile.Name,
		Description:         stepFile.Description,
		Website:             stepFile.Website,
		ForkURL:             stepFile.ForkURL,
		Source:              stepFile.Source,
		HostOsTags:          stepFile.HostOsTags,
		ProjectTypeTags:     stepFile.ProjectTypeTags,
		TypeTags:            stepFile.TypeTags,
		IsRequiresAdminUser: stepFile.IsRequiresAdminUser,
		Inputs:              inputs,
		Outputs:             outputs,
	}

	return step, nil
}

// ToWorkflowModel ...
func (workflowFile WorkflowFileModel) ToWorkflowModel() (WorkflowModel, error) {
	environments := []EnvironmentItemModel{}
	for _, envFile := range workflowFile.Environments {
		env, err := envFile.ToEnvironmentItemModel()
		if err != nil {
			return WorkflowModel{}, err
		}
		environments = append(environments, env)
	}

	steps := []StepListItem{}
	for _, stepListFile := range workflowFile.Steps {
		stepList := StepListItem{}
		for key, stepFile := range stepListFile {
			step, err := stepFile.ToStepModel()
			if err != nil {
				return WorkflowModel{}, err
			}
			stepList[key] = step
		}
		steps = append(steps, stepList)
	}

	worflow := WorkflowModel{
		Environments: environments,
		Steps:        steps,
	}

	return worflow, nil
}

// ToAppModel ...
func (appFile AppFileModel) ToAppModel() (AppModel, error) {
	environments := []EnvironmentItemModel{}
	for _, envFile := range appFile.Environments {
		env, err := envFile.ToEnvironmentItemModel()
		if err != nil {
			return AppModel{}, err
		}
		environments = append(environments, env)
	}

	app := AppModel{
		Environments: environments,
	}

	return app, nil
}

// ToBitriseConfigModel ...
func (configFile BitriseConfigFileModel) ToBitriseConfigModel() (BitriseConfigModel, error) {
	workflows := map[string]WorkflowModel{}
	for key, workflowFile := range configFile.Workflows {
		workfow, err := workflowFile.ToWorkflowModel()
		if err != nil {
			return BitriseConfigModel{}, err
		}
		workflows[key] = workfow
	}

	app, err := configFile.App.ToAppModel()
	if err != nil {
		return BitriseConfigModel{}, err
	}

	config := BitriseConfigModel{
		FormatVersion: configFile.FormatVersion,
		App:           app,
		Workflows:     workflows,
	}

	return config, nil
}
