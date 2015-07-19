package models

import (
	"testing"

	"gopkg.in/yaml.v2"
)

var (
	defaultTrue  = true
	defaultFalse = false
)

func TestMergeWith(t *testing.T) {
	stepData := StepModel{
		Name:        "name 1",
		Description: "desc 1",
		Website:     "web/1",
		ForkURL:     "fork/1",
		Source: StepSourceModel{
			Git: "https://git.url",
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: true,
		Inputs: []EnvironmentItemModel{
			EnvironmentItemModel{
				EnvKey: "KEY_1",
				Value:  "Value 1",
			},
			EnvironmentItemModel{
				EnvKey: "KEY_2",
				Value:  "Value 2",
			},
		},
		Outputs: []EnvironmentItemModel{},
	}
	stepDiffToMerge := StepModel{
		Name: "name 2",
		Inputs: []EnvironmentItemModel{
			EnvironmentItemModel{
				EnvKey: "KEY_2",
				Value:  "Value 2 CHANGED",
			},
		},
	}

	t.Logf("-> stepData: %#v\n", stepData)
	t.Logf("-> stepDiffToMerge: %#v\n", stepDiffToMerge)

	if err := stepData.MergeWith(stepDiffToMerge); err != nil {
		t.Error("Failed to convert: ", err)
	}

	t.Logf("-> MERGED Step Data: %#v\n", stepData)

	if stepData.Name != "name 2" {
		t.Error("step.Name incorrectly converted")
	}

	//
	t.Logf("-> MERGED Step Inputs: %#v\n", stepData.Inputs)
	if stepData.Inputs[0].EnvKey != "KEY_1" {
		t.Error("Inputs[0].EnvKey incorrectly converted")
	}
	if stepData.Inputs[0].Value != "Value 1" {
		t.Error("Inputs[0].Value incorrectly converted")
	}
	if stepData.Inputs[1].EnvKey != "KEY_2" {
		t.Error("Inputs[1].EnvKey incorrectly converted")
	}
	if stepData.Inputs[1].Value != "Value 2 CHANGED" {
		t.Error("Inputs[1].Value incorrectly converted")
	}
}

func TestToEnvironmentItemModel(t *testing.T) {
	envConf := EnvironmentItemSerializeModel{
		"TEST_KEY_1": "Test value 1",
		"opts": EnvironmentItemOptionsSerializeModel{
			Title:             "env title",
			Description:       "env description",
			ValueOptions:      []string{"one", "two"},
			IsRequired:        &defaultTrue,
			IsExpand:          &defaultFalse,
			IsDontChangeValue: &defaultTrue,
		},
	}

	envData, err := envConf.ToEnvironmentItemModel()
	if err != nil {
		t.Error("Failed to convert: ", err)
	}
	t.Logf("envData: %#v\n", envData)

	if envData.EnvKey != "TEST_KEY_1" {
		t.Error("envData.KEY incorrectly converted")
	}
	if envData.Value != "Test value 1" {
		t.Error("envData.Value incorrectly converted")
	}
	if envData.Title != "env title" {
		t.Error("envData.Title incorrectly converted")
	}
	if envData.Description != "env description" {
		t.Error("envData.Description incorrectly converted")
	}
	if len(envData.ValueOptions) != 2 {
		t.Error("envData.ValueOptions incorrectly converted")
	}
	if envData.IsRequired != true {
		t.Error("envData.IsRequired incorrectly converted")
	}
	if envData.IsExpand != false {
		t.Error("envData.IsExpand incorrectly converted")
	}
	if envData.IsDontChangeValue != true {
		t.Error("envData.IsDontChangeValue incorrectly converted")
	}
}
func TestToStepModel(t *testing.T) {
	confModel := StepSerializeModel{
		Name:        "test-step",
		Description: "test description",
		Website:     "https://web.site",
		ForkURL:     "https://fork.url",
		Source: StepSourceModel{
			Git: "https://git/url",
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"some-cat"},
		IsRequiresAdminUser: true,
		Inputs: []EnvironmentItemSerializeModel{
			EnvironmentItemSerializeModel{
				"INPUT_1": "Input value 1",
				"opts": EnvironmentItemOptionsSerializeModel{
					Title:       "Env title",
					Description: "Env description",
				},
			},
			EnvironmentItemSerializeModel{
				"INPUT_2": "Input value 2",
			},
		},
		Outputs: []EnvironmentItemSerializeModel{},
	}

	dataModel, err := confModel.ToStepModel()
	if err != nil {
		t.Error("Failed to convert: ", err)
	}
	t.Logf("dataModel: %#v\n", dataModel)

	if dataModel.Name != "test-step" {
		t.Error("dataModel.Name incorrectly converted")
	}
	if dataModel.Website != "https://web.site" {
		t.Error("dataModel.Website incorrectly converted")
	}
	if dataModel.IsRequiresAdminUser != true {
		t.Error("dataModel.IsRequiresAdminUser incorrectly converted")
	}

	inputs := dataModel.Inputs
	t.Logf("inputs: %#v\n", inputs)
	if inputs[0].EnvKey != "INPUT_1" {
		t.Error("inputs[0].EnvKey incorrectly converted")
	}
	if inputs[0].Value != "Input value 1" {
		t.Error("inputs[0].Value incorrectly converted")
	}
	if inputs[0].Title != "Env title" {
		t.Error("inputs[0].Title incorrectly converted")
	}
	if inputs[0].Description != "Env description" {
		t.Error("inputs[0].Description incorrectly converted")
	}

	if inputs[1].EnvKey != "INPUT_2" {
		t.Error("inputs[1].EnvKey incorrectly converted")
	}
	if inputs[1].Value != "Input value 2" {
		t.Error("inputs[1].Value incorrectly converted")
	}
}

func createTestBitriseConfigSerializeModel() BitriseConfigSerializeModel {
	confModel := BitriseConfigSerializeModel{
		FormatVersion: "0.0.1",
		App: AppSerializeModel{
			Environments: []EnvironmentItemSerializeModel{
				EnvironmentItemSerializeModel{
					"APP_KEY1": "App key 1",
					"opts": EnvironmentItemOptionsSerializeModel{
						Title:        "title",
						Description:  "descr",
						ValueOptions: []string{"1tes", "w"},
						IsRequired:   &defaultTrue,
						IsExpand:     &defaultFalse,
					},
				},
			},
		},
		Workflows: map[string]WorkflowSerializeModel{
			"test": WorkflowSerializeModel{
				Environments: []EnvironmentItemSerializeModel{},
				Steps: []StepListItemSerializeModel{
					StepListItemSerializeModel{
						"https://git/url::step-id@1.2.3": StepSerializeModel{
							Name:        "test-step",
							Description: "test description",
							Website:     "https://web.site",
							ForkURL:     "https://fork.url",
							Source: StepSourceModel{
								Git: "https://git/url",
							},
							HostOsTags:          []string{"osx"},
							ProjectTypeTags:     []string{"ios"},
							TypeTags:            []string{"some-cat"},
							IsRequiresAdminUser: true,
							Inputs: []EnvironmentItemSerializeModel{
								EnvironmentItemSerializeModel{
									"INPUT_KEY1": "Input key 1",
									"opts": EnvironmentItemOptionsSerializeModel{
										Title:        "input title 1",
										Description:  "input descr 1",
										ValueOptions: []string{"one", "two"},
										IsRequired:   &defaultTrue,
										IsExpand:     &defaultFalse,
									},
								},
							},
							Outputs: []EnvironmentItemSerializeModel{},
						},
					},
				},
			},
		},
	}
	return confModel
}

func TestConvertBitriseConfig(t *testing.T) {
	confModel := createTestBitriseConfigSerializeModel()

	dataModel, err := confModel.ToBitriseDataModel()
	if err != nil {
		t.Error("Failed to convert Bitrise Config model to Data model: ", err)
	}

	if dataModel.FormatVersion != "0.0.1" {
		t.Errorf("Format incorrectly converted to: %#v\n", dataModel.FormatVersion)
	}

	appEnv := dataModel.App.Environments[0]
	if appEnv.EnvKey != "APP_KEY1" {
		t.Errorf("App.Environments[0] (Key) incorrectly converted to: %#v\n", appEnv)
	}
	if appEnv.Value != "App key 1" {
		t.Errorf("App.Environments[0] (Value) incorrectly converted to: %#v\n", appEnv)
	}

	step := dataModel.Workflows["test"].Steps[0]
	stepID, stepData, err := step.GetStepIDStepDataPair()
	if err != nil {
		t.Logf("Step Data model: %#v\n", step)
		t.Error("Failed to convert Step Config model to Data model: ", err)
	}
	if stepID != "https://git/url::step-id@1.2.3" {
		t.Errorf("Workflows.Steps (StepID) incorrectly converted to: %#v\n", stepID)
	}
	t.Logf("StepData: %#v", stepData)
	if stepData.Name != "test-step" {
		t.Error("StepData (Name) incorrectly converted")
	}

	// Serialize & Deserialize
	confForSaveModel := dataModel.ToBitriseConfigSerializeModel()
	var bytes []byte
	bytes, err = yaml.Marshal(confForSaveModel)
	if err != nil {
		t.Error("Failed to generate YAML for Bitrise Config: ", err)
	}
	// deserialize
	var bitriseConfigFile BitriseConfigSerializeModel
	if err := yaml.Unmarshal(bytes, &bitriseConfigFile); err != nil {
		t.Error("Failed to parse YAML of Bitrise Config: ", err)
	}
	t.Logf("bitriseConfigFile: %#v\n", bitriseConfigFile)
	// finally, convert back to non-serialize data model
	finalBitriseDataModel, err := bitriseConfigFile.ToBitriseDataModel()
	if err != nil {
		t.Error("Failed to convert Bitrise serialize model to data model: ", err)
	}
	t.Logf("finalBitriseDataModel: %#v\n", finalBitriseDataModel)
}

func TestParseFromInterfaceMap(t *testing.T) {
	t.Log("EnvironmentItemOptionsSerializeModel::TestParseFromInterfaceMap - IMPLEMENT!")
}
