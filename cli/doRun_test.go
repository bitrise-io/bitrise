package cli

import (
	"strings"
	"testing"

	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	buildFailedTestWorkflowName      = "build_failed_test"
	buildFailedTestBitriseConfigPath = "./_tests/build_failed_test_bitrise.yml"
)

func TestBuildFailedMode(t *testing.T) {
	title := "title"

	beforeWorkflow := models.WorkflowModel{
		Steps: []models.StepListItemModel{
			"script": stepmanModels.StepModel{
				Title: strings.new("before script 1"),
			},
			"befor_script2": stepmanModels.StepModel{},
			"befor_script3": stepmanModels.StepModel{},
		},
	}

	stepRunResults := activateAndRunSteps(workflow, bitriseConfig.DefaultStepLibSource)
	buildRunResults.Append(stepRunResults)
}
