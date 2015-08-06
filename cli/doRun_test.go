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

func TestBuildFailedMode(t *testing.T) {
	config := getTestBitriseConfig()
	workflow, exist := config.Workflows["build_failed_test"]
	if !exist {
		t.Fatal("Failed to find workflow (build_failed_test) in config")
	}
	if workflow.Title == "" {
		workflow.Title = "build_failed_test"
	}

	buildRunResults := activateAndRunWorkflow(workflow, config, time.Now())
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

func getTestBitriseConfig() (config models.BitriseDataModel) {
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
		DefaultStepLibSource: "https://bitbucket.org/bitrise-team/bitrise-new-steps-spec",
		Workflows: map[string]models.WorkflowModel{
			"build_failed_test": targetWorkflow,
			"before1":           beforeWorkflow1,
			"before2":           beforeWorkflow2,
		},
	}

	return
}
