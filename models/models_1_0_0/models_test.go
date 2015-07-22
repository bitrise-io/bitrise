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
		t.Fatal("Failed to convert: ", err)
	}

	t.Logf("-> MERGED Step Data: %#v\n", stepData)

	if *stepData.Title != "name 2" {
		t.Fatal("step.Name incorrectly converted")
	}

	//
	t.Logf("-> MERGED Step Inputs: %#v\n", stepData.Inputs)
	input0 := stepData.Inputs[0]
	key0, value0, err := input0.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get key-value:", err)
	}
	if key0 != "KEY_1" {
		t.Fatal("Inputs[0].EnvKey incorrectly converted")
	}
	if value0 != "Value 1" {
		t.Fatal("Inputs[0].Value incorrectly converted")
	}

	input1 := stepData.Inputs[1]
	key1, value1, err := input1.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get key-value:", err)
	}
	if key1 != "KEY_2" {
		t.Fatal("Inputs[1].EnvKey incorrectly converted")
	}
	if value1 != "Value 2 CHANGED" {
		t.Fatal("Inputs[1].Value incorrectly converted")
	}
}

func TestParseFromInterfaceMap(t *testing.T) {
	t.Logf("TestParseFromInterfaceMap -- coming soon")
}

func TestCreateStepIDDataFromString(t *testing.T) {
	t.Logf("CreateStepIDDataFromString")

	// default / long / verbose ID mode
	stepCompositeIDString := "steplib-src::step-id@0.0.1"
	t.Log("stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err := CreateStepIDDataFromString(stepCompositeIDString, "")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.ID != "step-id" {
		t.Fatal("stepIDData.ID incorrectly converted:", stepIDData.ID)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// no steplib-source
	stepCompositeIDString = "step-id@0.0.1"
	t.Log("(no steplib-source, but default provided) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "default-steplib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "default-steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.ID != "step-id" {
		t.Fatal("stepIDData.ID incorrectly converted:", stepIDData.ID)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// invalid/empty step lib source, but default provided
	stepCompositeIDString = "::step-id@0.0.1"
	t.Log("(invalid/empty steplib source, but default provided) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "default-steplib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "default-steplib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.ID != "step-id" {
		t.Fatal("stepIDData.ID incorrectly converted:", stepIDData.ID)
	}
	if stepIDData.Version != "0.0.1" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// invalid/empty step lib source + no default
	stepCompositeIDString = "::step-id@0.0.1"
	t.Log("(invalid/empty steplib source) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err == nil {
		t.Fatal("Should fail to parse the ID if it contains an empty steplib-src and no default src is provided")
	}

	// no steplib-source & no default -> fail
	stepCompositeIDString = "step-id@0.0.1"
	t.Log("(no steplib-source & no default, should fail) stepCompositeIDString: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "")
	if err == nil {
		t.Fatal("Should fail to parse the ID if it does not contain a steplib-src and no default src is provided")
	} else {
		t.Log("Expected error (ok): ", err)
	}

	// no steplib & no version, only step-id
	stepCompositeIDString = "step-id"
	t.Log("no steplib & no version, only step-id and default lib source: ", stepCompositeIDString)
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-lib-src")
	if err != nil {
		t.Fatal("Failed to create StepIDData from composite-id: ", stepCompositeIDString, "| err:", err)
	}
	t.Logf("stepIDData:%#v", stepIDData)
	if stepIDData.SteplibSource != "def-lib-src" {
		t.Fatal("stepIDData.SteplibSource incorrectly converted:", stepIDData.SteplibSource)
	}
	if stepIDData.ID != "step-id" {
		t.Fatal("stepIDData.ID incorrectly converted:", stepIDData.ID)
	}
	if stepIDData.Version != "" {
		t.Fatal("stepIDData.Version incorrectly converted:", stepIDData.Version)
	}

	// empty test
	stepCompositeIDString = ""
	t.Log("Empty stepCompositeIDString test")
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-step-src")
	if err == nil {
		t.Fatal("Should fail to parse the ID from an empty string! (at least the step-id is required)")
	} else {
		t.Log("Expected error (ok): ", err)
	}

	// special empty test
	stepCompositeIDString = "@1.0.0"
	t.Log("Empty stepCompositeIDString test with only version")
	stepIDData, err = CreateStepIDDataFromString(stepCompositeIDString, "def-step-src")
	if err == nil {
		t.Fatal("Should fail to parse the ID from an empty string! (at least the step-id is required)")
	} else {
		t.Log("Expected error (ok): ", err)
	}
}
