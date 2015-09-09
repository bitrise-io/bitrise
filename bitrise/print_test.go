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
	PrintRunningStep("", "", 0)
	PrintRunningStep(longStr, "", 0)
	PrintRunningStep("", longStr, 0)
	PrintRunningStep(longStr, longStr, 0)
}

func TestGetTrimmedStepName(t *testing.T) {
	result := models.StepRunResultsModel{
		StepName: longStr,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	stepName := getTrimmedStepName(result)
	require.Equal(t, "This is a very long string,\nthis is a very ...", stepName)

	result = models.StepRunResultsModel{
		StepName: "",
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	stepName = getTrimmedStepName(result)
	require.Equal(t, "", stepName)
}

func TestStepResultCell(t *testing.T) {
	result := models.StepRunResultsModel{
		StepName: longStr,
		Status:   models.StepRunStatusCodeFailed,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}

	cell := stepResultCell(result)
	require.Equal(t, "| 🚫  | \x1b[31;1mThis is a very long string,\n... (exit code: 1)\x1b[0m| 0.01 sec |", cell)

	result = models.StepRunResultsModel{
		StepName: "",
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	cell = stepResultCell(result)
	require.Equal(t, "| ✅  | \x1b[32;1m\x1b[0m                                              | 0.00 sec |", cell)
}

func TestPrintStepSummary(t *testing.T) {
	result := models.StepRunResultsModel{
		StepName: longStr,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}
	PrintStepSummary(result, true)
	PrintStepSummary(result, false)

	result = models.StepRunResultsModel{
		StepName: "",
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

	result1 := models.StepRunResultsModel{
		StepName: longStr,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}
	result2 := models.StepRunResultsModel{
		StepName: "",
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

	result1 := models.StepRunResultsModel{
		StepName: longStr,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		Error:    errors.New(longStr),
		ExitCode: 1,
	}
	result2 := models.StepRunResultsModel{
		StepName: "",
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}
	PrintStepStatusList(longStr, []models.StepRunResultsModel{result1, result2})
}
