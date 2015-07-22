package models

import (
	"testing"

	stepmanModels "github.com/bitrise-io/stepman/models"
)

func TestMergeWith(t *testing.T) {
	title := "name 1"
	desc := "desc 1"
	website := "web/1"
	git := "https://git.url"
	fork := "fork/1"

	defaultTrue := true

	stepData := stepmanModels.StepModel{
		Title:         &title,
		Description:   &desc,
		Website:       &website,
		SourceCodeURL: &fork,
		Source: stepmanModels.StepSourceModel{
			Git: &git,
		},
		HostOsTags:          []string{"osx"},
		ProjectTypeTags:     []string{"ios"},
		TypeTags:            []string{"test"},
		IsRequiresAdminUser: &defaultTrue,
		Inputs: []stepmanModels.EnvironmentItemModel{
			stepmanModels.EnvironmentItemModel{
				"KEY_1": "Value 1",
			},
			stepmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2",
			},
		},
		Outputs: []stepmanModels.EnvironmentItemModel{},
	}

	diffTitle := "name 2"
	stepDiffToMerge := stepmanModels.StepModel{
		Title: &diffTitle,
		Inputs: []stepmanModels.EnvironmentItemModel{
			stepmanModels.EnvironmentItemModel{
				"KEY_2": "Value 2 CHANGED",
			},
		},
	}

	t.Logf("-> stepData: %#v\n", stepData)
	t.Logf("-> stepDiffToMerge: %#v\n", stepDiffToMerge)

	if err := MergeStepWith(stepData, stepDiffToMerge); err != nil {
		t.Error("Failed to convert: ", err)
	}

	t.Logf("-> MERGED Step Data: %#v\n", stepData)

	if *stepData.Title != "name 2" {
		t.Error("step.Name incorrectly converted")
	}

	//
	t.Logf("-> MERGED Step Inputs: %#v\n", stepData.Inputs)
	input0 := stepData.Inputs[0]
	key0, value0, err := input0.GetKeyValuePair()
	if err != nil {
		t.Error("Failed to get key-value:", err)
	}
	if key0 != "KEY_1" {
		t.Error("Inputs[0].EnvKey incorrectly converted")
	}
	if value0 != "Value 1" {
		t.Error("Inputs[0].Value incorrectly converted")
	}

	input1 := stepData.Inputs[1]
	key1, value1, err := input1.GetKeyValuePair()
	if err != nil {
		t.Error("Failed to get key-value:", err)
	}
	if key1 != "KEY_2" {
		t.Error("Inputs[1].EnvKey incorrectly converted")
	}
	if value1 != "Value 2 CHANGED" {
		t.Error("Inputs[1].Value incorrectly converted")
	}
}

func TestParseFromInterfaceMap(t *testing.T) {
	t.Logf("TestParseFromInterfaceMap -- coming soon")
}
