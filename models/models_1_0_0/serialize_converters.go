package models

// ToStepModel ...
func (stepFile StepSerializeModel) ToStepModel() (StepModel, error) {
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
		Name:                stepFile.Name,
		Description:         stepFile.Description,
		Website:             stepFile.Website,
		ForkURL:             stepFile.ForkURL,
		Source:              stepFile.Source,
		HostOsTags:          stepFile.HostOsTags,
		ProjectTypeTags:     stepFile.ProjectTypeTags,
		TypeTags:            stepFile.TypeTags,
		IsRequiresAdminUser: stepFile.IsRequiresAdminUser,
		IsAlwaysRun:         stepFile.IsAlwaysRun,
		Inputs:              inputs,
		Outputs:             outputs,
	}

	return step, nil
}

// ToWorkflowModel ...
func (workflowFile WorkflowSerializeModel) ToWorkflowModel() (WorkflowModel, error) {
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
func (appConfig AppSerializeModel) ToAppModel() (AppModel, error) {
	environments := []EnvironmentItemModel{}
	for _, envFile := range appConfig.Environments {
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

// ToBitriseDataModel ...
func (confModel BitriseConfigSerializeModel) ToBitriseDataModel() (BitriseDataModel, error) {
	workflows := map[string]WorkflowModel{}
	for key, workflowFile := range confModel.Workflows {
		workfow, err := workflowFile.ToWorkflowModel()
		if err != nil {
			return BitriseDataModel{}, err
		}
		workflows[key] = workfow
	}

	app, err := confModel.App.ToAppModel()
	if err != nil {
		return BitriseDataModel{}, err
	}

	config := BitriseDataModel{
		FormatVersion: confModel.FormatVersion,
		App:           app,
		Workflows:     workflows,
	}

	return config, nil
}

// ToWorkflowSerializeModel ...
func (wfModel WorkflowModel) ToWorkflowSerializeModel() WorkflowSerializeModel {
	// // WorkflowSerializeModel ...
	// type WorkflowSerializeModel struct {
	// 	Environments []EnvironmentItemSerializeModel `json:"environments"`
	// 	Steps        []StepListItemSerializeModel    `json:"steps"`
	// }
	//
	// type StepListItemSerializeModel map[string]StepSerializeModel

	environments := []EnvironmentItemSerializeModel{}
	for _, env := range wfModel.Environments {
		environments = append(environments, env.ToEnvironmentItemSerializeModel())
	}

	steps := []StepListItemSerializeModel{}
	for _, stepListFile := range wfModel.Steps {
		stepList := StepListItemSerializeModel{}
		for key, aStep := range stepListFile {
			stepList[key] = aStep.ToStepSerializeModel()
		}
		steps = append(steps, stepList)
	}

	worflow := WorkflowSerializeModel{
		Environments: environments,
		Steps:        steps,
	}

	return worflow
}

// ToEnvironmentItemModel ...
func (envFile EnvironmentItemSerializeModel) ToEnvironmentItemModel() (EnvironmentItemModel, error) {
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
		EnvKey:            key,
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

// ToEnvironmentItemSerializeModel ...
func (envItm EnvironmentItemModel) ToEnvironmentItemSerializeModel() EnvironmentItemSerializeModel {
	return EnvironmentItemSerializeModel{
		envItm.EnvKey: envItm.Value,
		OptionsKey: EnvironmentItemOptionsSerializeModel{
			Title:             envItm.Title,
			Description:       envItm.Description,
			ValueOptions:      envItm.ValueOptions,
			IsRequired:        &envItm.IsRequired,
			IsExpand:          &envItm.IsExpand,
			IsDontChangeValue: &envItm.IsDontChangeValue,
		},
	}
}

// ToAppSerializeModel ...
func (appData AppModel) ToAppSerializeModel() AppSerializeModel {
	environments := []EnvironmentItemSerializeModel{}
	for _, envItm := range appData.Environments {
		environments = append(environments, envItm.ToEnvironmentItemSerializeModel())
	}

	app := AppSerializeModel{
		Environments: environments,
	}

	return app
}
