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
	expected := ""
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
}

func TestStatusReasonFailed(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodeFailed,
		ExitCode: 25,
	}
	expected := "exit code: 25"
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
}

func TestStatusReasonPreparationFailed(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status:   StepRunStatusCodePreparationFailed,
		ExitCode: 30,
	}
	expected := "exit code: 30"
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
}

func TestStatusReasonFailedSkippable(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: StepRunStatusCodeFailedSkippable,
	}
	expected := "This Step failed, but it was marked as “is_skippable”, so the build continued."
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
}

func TestStatusReasonSkipped(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: StepRunStatusCodeSkipped,
	}
	expected := "This Step was skipped, because a previous Step failed, and this Step was not marked “is_always_run”."
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
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
	expected := `This Step was skipped, because its “run_if” expression evaluated to false.

The “run_if” expression was: 2+2==4`
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
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
	}
	expected := "This Step timed out after 9h 50s."
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
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
	}
	expected := "This Step failed, because it has not sent any output for 32s."
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
}

func TestStatusReasonDefault(t *testing.T) {
	var s StepRunResultsModel = StepRunResultsModel{
		Status: -999,
	}
	expected := "unknown result code"
	actual := s.StatusReason()

	assert.Equal(t, expected, actual)
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
		StepRunStatusCodeSuccess:                "Success",
		StepRunStatusCodeFailed:                 "Failed",
		StepRunStatusCodeFailedSkippable:        "Failed",
		StepRunStatusCodeSkipped:                "Skipped",
		StepRunStatusCodeSkippedWithRunIf:       "Skipped",
		StepRunStatusCodePreparationFailed:      "Failed",
		StepRunStatusAbortedWithCustomTimeout:   "Failed",
		StepRunStatusAbortedWithNoOutputTimeout: "Failed",
	}
	actual := make(map[StepRunStatus]string)

	for k := range expected {
		var s StepRunResultsModel = StepRunResultsModel{
			Status: k,
		}
		actual[k] = s.ShortReason()
	}

	assert.Equal(t, expected, actual)
}
