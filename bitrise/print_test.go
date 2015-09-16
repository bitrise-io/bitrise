package bitrise

import (
	"errors"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/require"
)

const longStr = `This is a very long string,
this is a very long string,
this is a very long string,
this is a very long string,
this is a very long string,
this is a very long string.
`

func TestPrintRunningWorkflow(t *testing.T) {
	PrintRunningWorkflow(longStr)
}

func TestPrintRunningStep(t *testing.T) {
	stepInfo := models.StepInfoModel{
		ID:      "",
		Version: "",
	}
	PrintRunningStep(stepInfo, 0)

	stepInfo.ID = longStr
	stepInfo.Version = ""
	PrintRunningStep(stepInfo, 0)

	stepInfo.ID = ""
	stepInfo.Version = longStr
	PrintRunningStep(stepInfo, 0)

	stepInfo.ID = longStr
	stepInfo.Version = longStr
	PrintRunningStep(stepInfo, 0)
}

func TestGetTrimmedStepName(t *testing.T) {
	stepInfo := models.StepInfoModel{
		ID:      longStr,
		Version: longStr,
	}

	result := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	stepName := getTrimmedStepName(result)
	require.Equal(t, "This is a very ... (...s a very long string.\n)", stepName)

	stepInfo.ID = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	stepName = getTrimmedStepName(result)
	require.Equal(t, " (...s a very long string.\n)", stepName)
}

func TestStepResultCell(t *testing.T) {
	stepInfo := models.StepInfoModel{
		ID:      longStr,
		Version: longStr,
	}

	result := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeFailed,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	cell := stepResultCell(result)
	require.Equal(t, "| 🚫  | \x1b[31;1m... (...s a very long string.\n) (exit code: 1)\x1b[0m| 0.01 sec |", cell)

	stepInfo.ID = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	cell = stepResultCell(result)
	require.Equal(t, "| ✅  | \x1b[32;1m (...s a very long string.\n)\x1b[0m                  | 0.00 sec |", cell)
}

func TestPrintStepSummary(t *testing.T) {
	stepInfo := models.StepInfoModel{
		ID:      longStr,
		Version: longStr,
	}

	result := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}
	PrintStepSummary(result, true)
	PrintStepSummary(result, false)

	stepInfo.ID = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}
	PrintStepSummary(result, true)
	PrintStepSummary(result, false)
}

func TestPrintSummary(t *testing.T) {
	PrintSummary(models.BuildRunResultsModel{})

	stepInfo := models.StepInfoModel{
		ID:      longStr,
		Version: longStr,
	}

	result1 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	stepInfo.ID = ""
	result2 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	buildResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
		SuccessSteps:   []models.StepRunResultsModel{result1, result2},
	}

	PrintSummary(buildResults)
}

func TestPrintStepStatusList(t *testing.T) {
	PrintStepStatusList("", []models.StepRunResultsModel{})

	stepInfo := models.StepInfoModel{
		ID:      longStr,
		Version: longStr,
	}

	result1 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	stepInfo.ID = ""
	result2 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}
	PrintStepStatusList(longStr, []models.StepRunResultsModel{result1, result2})
}
