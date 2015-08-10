package cli

import (
	"os"
	"testing"
	"time"

	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/utils"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	buildFailedTestWorkflowName      = "build_failed_test"
	buildFailedTestBitriseConfigPath = "./_tests/build_failed_test_bitrise.yml"
)

func Test0Steps1Workflows(t *testing.T) {
	workflow := models.WorkflowModel{}

	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"zero_steps": models.WorkflowModel{},
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 0 {
		t.Fatalf("Success step count (%d), should be (0)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func Test0Steps3WorkflowsBeforeAfter(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

	beforeWorkflow := models.WorkflowModel{}
	afterWorkflow := models.WorkflowModel{}

	workflow := models.WorkflowModel{
		BeforeRun: []string{"before"},
		AfterRun:  []string{"after"},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target": workflow,
			"before": beforeWorkflow,
			"after":  afterWorkflow,
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 0 {
		t.Fatalf("Success step count (%d), should be (0)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func Test0Steps3WorkflowsCircularDependency(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

	beforeWorkflow := models.WorkflowModel{
		BeforeRun: []string{"target"},
	}

	afterWorkflow := models.WorkflowModel{}

	workflow := models.WorkflowModel{
		BeforeRun: []string{"before"},
		AfterRun:  []string{"after"},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target": workflow,
			"before": beforeWorkflow,
			"after":  afterWorkflow,
		},
	}

	if err := config.Validate(); err == nil {
		t.Fatal("Circular dependency, should fail")
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func Test1Workflow(t *testing.T) {
	defaultTrue := true
	shouldSuccess := "Should success"
	shouldFail := "Should fail"
	shouldFailButNotImporatnt := "Should failed not important"
	shouldSkipped := "Should skipp"

	workflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
				},
			},
			models.StepListItemModel{ // Failed, but not important
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldFailButNotImporatnt),
					IsSkippable: utils.NewBoolPtr(defaultTrue),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
									#!/bin/bash
									set -v
									echo 'This is a before workflow'
									exit 34
								`,
						},
					},
				},
			},
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
				},
			},
			models.StepListItemModel{ // Fail
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldFail),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
									#!/bin/bash
									set -v
									echo 'This is a before workflow'
									exit 34
								`,
						},
					},
				},
			},
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldSuccess),
					IsAlwaysRun: utils.NewBoolPtr(defaultTrue),
				},
			},
			models.StepListItemModel{ // Skipped
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSkipped),
				},
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"trivial_fail": workflow,
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 1 {
		t.Fatalf("FailedNotImportant step count (%d), should be (1)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func Test3Workflows(t *testing.T) {
	// Before
	beforeWorkflow1 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{},
			},
			models.StepListItemModel{ // Fail, not important
				"script": stepmanModels.StepModel{
					IsSkippable: utils.NewBoolPtr(true),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
									#!/bin/bash
									set -v
									echo 'Before step 1'
									exit 1
								`,
						},
					},
				},
			},
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{},
			},
		},
	}

	beforeWorkflow2 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{},
			},
		},
	}

	// Target
	workflow := models.WorkflowModel{
		BeforeRun: []string{"before1", "before2"},
		AfterRun:  []string{"after1", "after2"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Fail
				"script": stepmanModels.StepModel{
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
										#!/bin/bash
										set -v
										echo 'Target step'
										exit 1
									`,
						},
					},
				},
			},
		},
	}

	// After
	afterWorkflow1 := models.WorkflowModel{
		Steps: []models.StepListItemModel{ // Fail
			models.StepListItemModel{
				"script": stepmanModels.StepModel{
					IsAlwaysRun: utils.NewBoolPtr(true),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
											#!/bin/bash
											set -v
											exit 1'
										`,
						},
					},
				},
			},
		},
	}

	afterWorkflow2 := models.WorkflowModel{
		Steps: []models.StepListItemModel{ // Skipp
			models.StepListItemModel{
				"script": stepmanModels.StepModel{},
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target":  workflow,
			"before1": beforeWorkflow1,
			"before2": beforeWorkflow2,
			"after1":  afterWorkflow1,
			"after2":  afterWorkflow2,
		},
	}

	err := config.Validate()
	if err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 2 {
		t.Fatalf("Failed step count (%d), should be (2)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 1 {
		t.Fatalf("FailedNotImportant step count (%d), should be (1)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func TestMasterWorkflow(t *testing.T) {
	// Before
	beforeStep1 := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"content": `
					#!/bin/bash
					set -v
					echo 'Before step 1'
				`,
			},
		},
	}

	beforeWorkflow1 := models.WorkflowModel{
		BeforeRun: []string{"before2"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": beforeStep1,
			},
		},
	}

	beforeStep2 := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"content": `
					#!/bin/bash
					set -v
					echo 'Before step 2'
				`,
			},
		},
	}

	beforeWorkflow2 := models.WorkflowModel{
		BeforeRun: []string{"before1"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": beforeStep2,
			},
		},
	}

	// After
	afterStep := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"content": `
						#!/bin/bash
						set -v
						echo 'After step'
					`,
			},
		},
	}

	afterWorkflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": afterStep,
			},
		},
	}

	// Target
	targetStep := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"content": `
						#!/bin/bash
						set -v
						echo 'Target step'
						exit 1
					`,
			},
		},
	}

	workflow := models.WorkflowModel{
		BeforeRun: []string{"before1", "before2"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": targetStep,
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target":  workflow,
			"before1": beforeWorkflow1,
			"before2": beforeWorkflow2,
			"after":   afterWorkflow,
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("Should found workflow reference cycle")
	}
}

func TestBuildStatusEnv(t *testing.T) {
	defaultTrue := true
	shouldSuccess := "Should success"
	shouldFail := "Should fail"
	shouldFailButNotImporatnt := "Should failed not important"
	shouldSkipped := "Should skipp"

	workflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								if [[ "$BITRISE_BUILD_STATUS" != "0" ]] ; then
								  exit 1
								fi
								if [[ "$STEPLIB_BUILD_STATUS" != "0" ]] ; then
								  exit 1
								fi
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Failed, but not important
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldFailButNotImporatnt),
					IsSkippable: utils.NewBoolPtr(defaultTrue),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
									#!/bin/bash
									set -v
									echo 'This is a before workflow'
									exit 34
								`,
						},
					},
				},
			},
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								if [[ "$BITRISE_BUILD_STATUS" != "0" ]] ; then
								  exit 1
								fi
								if [[ "$STEPLIB_BUILD_STATUS" != "0" ]] ; then
								  exit 1
								fi
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Fail
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldFail),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								exit 1
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldSuccess),
					IsAlwaysRun: utils.NewBoolPtr(defaultTrue),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								if [[ "$BITRISE_BUILD_STATUS" != "1" ]] ; then
								  echo "should fail"
								fi
								if [[ "$STEPLIB_BUILD_STATUS" != "1" ]] ; then
								  echo "should fail"
								fi
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Skipped
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSkipped),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								if [[ "$BITRISE_BUILD_STATUS" != "1" ]] ; then
								  echo "should fail"
								fi
								if [[ "$STEPLIB_BUILD_STATUS" != "1" ]] ; then
								  echo "should fail"
								fi
							`,
						},
					},
				},
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"trivial_fail": workflow,
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	t.Logf("Build run results: %#v\n", buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 1 {
		t.Fatalf("FailedNotImportant step count (%d), should be (1)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func TestTivialFail(t *testing.T) {
	defaultTrue := true
	shouldSuccess := "Should success"
	shouldFail := "Should fail"
	shouldFailButNotImporatnt := "Should failed not important"
	shouldSkipped := "Should skipp"

	workflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Should be success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
				},
			},
			models.StepListItemModel{ // Should fail, but not important
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldFailButNotImporatnt),
					IsSkippable: utils.NewBoolPtr(defaultTrue),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								echo 'This is a before workflow'
								exit 34
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Should success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccess),
				},
			},
			models.StepListItemModel{ // Should fail
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldFail),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
								#!/bin/bash
								set -v
								echo 'This is a before workflow'
								exit 34
							`,
						},
					},
				},
			},
			models.StepListItemModel{ // Should be skipped
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSkipped),
				},
			},
			models.StepListItemModel{ // Should be success
				"script": stepmanModels.StepModel{
					Title:       utils.NewStringPtr(shouldSuccess),
					IsAlwaysRun: utils.NewBoolPtr(defaultTrue),
				},
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"trivial_fail": workflow,
		},
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	t.Log("Buil run results:", buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 1 {
		t.Fatalf("FailedNotImportant step count (%d), should be (1)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func TestTrivialSuccess(t *testing.T) {
	step := stepmanModels.StepModel{
		Inputs: []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"content": `
						#!/bin/bash
						set -v
						echo 'Should be success'
					`,
			},
		},
	}

	workflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": step,
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"trivial_success": workflow,
		},
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

func TestBuildFailedMode(t *testing.T) {
	shouldSuccessBefore1Before1 := "Should success (before1 - before1)"
	shouldFailBefore1Before2 := "Should fail (before1 - befor2)"
	shouldSkippedBefor2Befor1 := "Should skipp (before2 - before1)"
	shouldSkippedTargetStep1 := "Should skipp (target - step1)"

	// Before
	beforeWorkflow1 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Success
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSuccessBefore1Before1),
				},
			},
			models.StepListItemModel{ // Fail
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldFailBefore1Before2),
					Inputs: []envmanModels.EnvironmentItemModel{
						envmanModels.EnvironmentItemModel{
							"content": `
									#!/bin/bash
									set -v
									echo 'This is a before workflow'
									exit 34
								`,
						},
					},
				},
			},
		},
	}

	beforeWorkflow2 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Skip
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSkippedBefor2Befor1),
				},
			},
		},
	}

	// Target
	targetWorkflow := models.WorkflowModel{
		BeforeRun: []string{"before1", "before2"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{ // Skip
				"script": stepmanModels.StepModel{
					Title: utils.NewStringPtr(shouldSkippedTargetStep1),
				},
			},
		},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"build_failed_test": targetWorkflow,
			"before1":           beforeWorkflow1,
			"before2":           beforeWorkflow2,
		},
	}
	workflow, exist := config.Workflows["build_failed_test"]
	if !exist {
		t.Fatal("Failed to find workflow (build_failed_test) in config")
	}
	if workflow.Title == "" {
		workflow.Title = "build_failed_test"
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}

	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SkippedSteps) != 2 {
		t.Fatalf("Skipped step count (%d), should be (2)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}
