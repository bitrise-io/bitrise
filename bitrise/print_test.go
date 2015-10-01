package bitrise

import (
	"errors"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
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

func TestPrintRunningStepHeader(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		ID:      "",
		Version: "",
	}
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.ID = longStr
	stepInfo.Version = ""
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.ID = ""
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.ID = longStr
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, 0)
}

func TestGetTrimmedStepName(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
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
	require.Equal(t, "This is a very long string,\nth... (...s a very long string.\n)", stepName)

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

func TestgetRunningStepFooterMainSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
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

	cell := getRunningStepFooterMainSection(result)
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

	cell = getRunningStepFooterMainSection(result)
	require.Equal(t, "| ✅  | \x1b[32;1m (...s a very long string.\n)\x1b[0m                  | 0.00 sec |", cell)
}

func TestPrintRunningStepFooter(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
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
	PrintRunningStepFooter(result, true)
	PrintRunningStepFooter(result, false)

	stepInfo.ID = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}
	PrintRunningStepFooter(result, true)
	PrintRunningStepFooter(result, false)
}

func TestPrintSummary(t *testing.T) {
	PrintSummary(models.BuildRunResultsModel{})

	stepInfo := stepmanModels.StepInfoModel{
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

	stepInfo := stepmanModels.StepInfoModel{
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
