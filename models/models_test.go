package models

import (
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatusReasonSuccess(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: StepRunStatusCodeSuccess,
	}
	expectedStatusReason := ""
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonFailed(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodeFailed,
		ExitCode: 25,
		ErrorStr: "exit code: 25",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 25, Message: "exit code: 25"}}
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonPreparationFailed(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodePreparationFailed,
		ExitCode: 30,
		ErrorStr: "Failed to clone step.",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 30, Message: "Failed to clone step."}}
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonFailedSkippable(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodeFailedSkippable,
		ExitCode: 10,
		ErrorStr: "ABCD",
	}
	expectedStatusReason := "This Step failed, but it was marked as \"is_skippable\", so the build continued."
	expectedStepErrors := []StepError{{Code: 10, Message: "ABCD"}}
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonSkipped(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: StepRunStatusCodeSkipped,
	}
	expectedStatusReason := "This Step was skipped, because a previous Step failed, and this Step was not marked “is_always_run”."
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonSkippedWithRunIf(t *testing.T) {
	var runif string = "2+2==4"
	var model models.StepModel = models.StepModel{
		RunIf: &runif,
	}
	var info models.StepInfoModel = models.StepInfoModel{
		Step: model,
	}
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodeSkippedWithRunIf,
		StepInfo: info,
	}
	expectedStatusReason := `This Step was skipped, because its “run_if” expression evaluated to false.
The “run_if” expression was: 2+2==4`
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonCustomTimeout(t *testing.T) {
	var timeout int = 32450
	var model models.StepModel = models.StepModel{
		Timeout: &timeout,
	}
	var info models.StepInfoModel = models.StepInfoModel{
		Step: model,
	}
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusAbortedWithCustomTimeout,
		StepInfo: info,
		ExitCode: 5,
		ErrorStr: "This won't be used.",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 5, Message: "This Step timed out after 9h 50s."}}
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonNoOutputTimeout(t *testing.T) {
	var noOutputTimeout int = 32
	var model models.StepModel = models.StepModel{
		NoOutputTimeout: &noOutputTimeout,
	}
	var info models.StepInfoModel = models.StepInfoModel{
		Step: model,
	}
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusAbortedWithNoOutputTimeout,
		StepInfo: info,
		ExitCode: 6,
		ErrorStr: "This won't be used.",
	}
	expectedStatusReason := ""
	expectedStepErrors := []StepError{{Code: 6, Message: "This Step failed, because it has not sent any output for 32s."}}
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestStatusReasonDefault(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: -999,
	}
	expectedStatusReason := ""
	var expectedStepErrors []StepError
	actualStatusReason, actualStepErrors := s.StatusReasons()

	assert.Equal(t, expectedStatusReason, actualStatusReason)
	assert.Equal(t, expectedStepErrors, actualStepErrors)
}

func TestFormatStatusReasonTimeInterval(t *testing.T) {
	expected := map[int]string{
		10:   "10s",
		60:   "1m",
		61:   "1m 1s",
		3600: "1h",
		3601: "1h 1s",
		3661: "1h 1m 1s",
	}

	actual := make(map[int]string)

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
		var s StepRunResultsModel = StepRunResultsModel{
			Status: k,
		}
		actual[k] = s.Status.Name()
	}

	assert.Equal(t, expected, actual)
}
