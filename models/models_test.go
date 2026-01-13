package models

import (
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/models/yml"
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestStatusReasonSuccess(t *testing.T) {
	var s = StepRunResultsModel{
		Status: StepRunStatusCodeSuccess,
	}
	expectedStatusReason := ""
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonFailed(t *testing.T) {
	var s = StepRunResultsModel{
		Status:   StepRunStatusCodeFailed,
		ExitCode: 25,
		ErrorStr: "exit code: 25",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 25, Message: "exit code: 25"}}
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonPreparationFailed(t *testing.T) {
	var s = StepRunResultsModel{
		Status:   StepRunStatusCodePreparationFailed,
		ExitCode: 30,
		ErrorStr: "Failed to clone step.",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 30, Message: "Failed to clone step."}}
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonFailedSkippable(t *testing.T) {
	var s = StepRunResultsModel{
		Status:   StepRunStatusCodeFailedSkippable,
		ExitCode: 10,
		ErrorStr: "ABCD",
	}
	expectedStatusReason := "This Step failed, but it was marked as \"is_skippable\", so the build continued."
	expectedStepErrors := []StepError{{Code: 10, Message: "ABCD"}}
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonSkipped(t *testing.T) {
	var s = StepRunResultsModel{
		Status: StepRunStatusCodeSkipped,
	}
	expectedStatusReason := `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonSkippedWithRunIf(t *testing.T) {
	var runif = "2+2==4"
	var model = models.StepModel{
		RunIf: &runif,
	}
	var info = models.StepInfoModel{
		Step: model,
	}
	var s = StepRunResultsModel{
		Status:   StepRunStatusCodeSkippedWithRunIf,
		StepInfo: info,
	}
	expectedStatusReason := `This Step was skipped, because its "run_if" expression evaluated to false.
The "run_if" expression was: 2+2==4`
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonCustomTimeout(t *testing.T) {
	timeout := 32450 * time.Second
	model := models.StepModel{}
	info := models.StepInfoModel{
		Step: model,
	}
	var s = StepRunResultsModel{
		Status:   StepRunStatusAbortedWithCustomTimeout,
		StepInfo: info,
		ExitCode: 5,
		ErrorStr: "This won't be used.",
		Timeout:  timeout,
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 5, Message: "This Step timed out after 9h 50s."}}
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonNoOutputTimeout(t *testing.T) {
	noOutputTimeout := 32 * time.Second
	model := models.StepModel{}
	info := models.StepInfoModel{
		Step: model,
	}
	var s = StepRunResultsModel{
		Status:          StepRunStatusAbortedWithNoOutputTimeout,
		StepInfo:        info,
		ExitCode:        6,
		ErrorStr:        "This won't be used.",
		NoOutputTimeout: noOutputTimeout,
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 6, Message: "This Step failed, because it has not sent any output for 32s."}}
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonDefault(t *testing.T) {
	var s = StepRunResultsModel{
		Status: -999,
	}
	expectedStatusReason := ""
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasonAndErrors()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestFormatStatusReasonTimeInterval(t *testing.T) {
	expected := map[time.Duration]string{
		10 * time.Second:   "10s",
		60 * time.Second:   "1m",
		61 * time.Second:   "1m 1s",
		3600 * time.Second: "1h",
		3601 * time.Second: "1h 1s",
		3661 * time.Second: "1h 1m 1s",
	}

	actual := make(map[time.Duration]string)

	for key := range expected {
		actual[key] = formatStatusReasonTimeInterval(key)
	}

	assert.Equal(t, expected, actual)
}

func TestShortReason(t *testing.T) {
	expected := map[StepRunStatus]string{
		StepRunStatusCodeSuccess:                "",
		StepRunStatusCodeFailed:                 "Failed",
		StepRunStatusCodeFailedSkippable:        "Failed",
		StepRunStatusCodeSkipped:                "Skipped",
		StepRunStatusCodeSkippedWithRunIf:       "Skipped",
		StepRunStatusCodePreparationFailed:      "Failed",
		StepRunStatusAbortedWithCustomTimeout:   "Failed",
		StepRunStatusAbortedWithNoOutputTimeout: "Failed",
		-999:                                    "", //default case
	}
	actual := make(map[StepRunStatus]string)

	for k := range expected {
		var s = StepRunResultsModel{
			Status: k,
		}
		actual[k] = s.Status.Name()
	}

	assert.Equal(t, expected, actual)
}

func TestGraphPipelineWorkflow(t *testing.T) {
	testCases := []struct {
		rawYML        string
		errorExpected bool
	}{
		{
			rawYML: `
pipelines:
  pipeline1:
    workflows:
      workflow1:
        abort_on_fail: true
        should_always_run: off
        run_if:
          expression: "custom-expression"
workflows:
  workflow1: {}
`,
		},
		{
			rawYML: `
pipelines:
  pipeline1:
    workflows:
      workflow1:
        abort_on_fail: false
        should_always_run: workflow
workflows:
  workflow1: {}
`,
		},
		{
			rawYML: `
pipelines:
  pipeline1:
    workflows:
      workflow1:
        should_always_run: none
workflows:
  workflow1: {}
`,
			errorExpected: true,
		},
	}

	for _, testCase := range testCases {
		config := yml.BitriseDataModel{}
		err := yaml.Unmarshal([]byte(testCase.rawYML), &config)
		if testCase.errorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
