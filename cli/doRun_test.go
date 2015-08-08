package cli

import (
	"testing"
	"time"

	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	buildFailedTestWorkflowName      = "build_failed_test"
	buildFailedTestBitriseConfigPath = "./_tests/build_failed_test_bitrise.yml"
)

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

func TestZeroSteps(t *testing.T) {
	workflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"zero_steps": models.WorkflowModel{},
		},
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
}

func TestTivialFail(t *testing.T) {
	step := stepmanModels.StepModel{
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
			"trivial_fail": workflow,
		},
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	if len(buildRunResults.SuccessSteps) != 0 {
		t.Fatalf("Success step count (%d), should be (0)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
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
}

func TestBuildFailedMode(t *testing.T) {
	config := getBuildFailedTestBitriseConfig()
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
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedNotImportantSteps) != 0 {
		t.Fatalf("FailedNotImportant step count (%d), should be (0)", len(buildRunResults.FailedNotImportantSteps))
	}
	if len(buildRunResults.SkippedSteps) != 2 {
		t.Fatalf("Skipped step count (%d), should be (2)", len(buildRunResults.SkippedSteps))
	}
}

func getBuildFailedTestBitriseConfig() (config models.BitriseDataModel) {
	// Before
	beforeStep1 := stepmanModels.StepModel{
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
	}

	beforeWorkflow1 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": stepmanModels.StepModel{},
			},
			models.StepListItemModel{
				"script": beforeStep1,
			},
		},
	}

	beforeWorkflow2 := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": stepmanModels.StepModel{},
			},
		},
	}

	// Target
	targetWorkflow := models.WorkflowModel{
		BeforeRun: []string{"before1", "before2"},
		Steps: []models.StepListItemModel{
			models.StepListItemModel{
				"script": stepmanModels.StepModel{},
			},
		},
	}

	config = models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"build_failed_test": targetWorkflow,
			"before1":           beforeWorkflow1,
			"before2":           beforeWorkflow2,
		},
	}

	return
}
