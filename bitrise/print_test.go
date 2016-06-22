package bitrise

import (
	"errors"
	"strings"
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

func TestIsUpdateAvailable(t *testing.T) {
	t.Log("simple compare versions - ture")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version: "1.0.0",
			Latest:  "1.1.0",
		}

		require.Equal(t, true, isUpdateAvailable(stepInfo1))
	}

	t.Log("simple compare versions - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version: "1.0.0",
			Latest:  "1.0.0",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}

	t.Log("issue - no latest - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version: "1.0.0",
			Latest:  "",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}

	t.Log("issue - no current - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version: "",
			Latest:  "1.0.0",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}
}

func TestGetTrimmedStepName(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Title:   longStr,
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
	require.Equal(t, "This is a very long string,\nthis is a very long string,\nth...", stepName)

	stepInfo.Title = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	stepName = getTrimmedStepName(result)
	require.Equal(t, "", stepName)
}

func TestGetRunningStepHeaderMainSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Title:   longStr,
		Version: longStr,
	}

	actual := getRunningStepHeaderMainSection(stepInfo, 0)
	expected := "| (0) This is a very long string,\nthis is a very long string,\nthis is a ver... |"
	require.Equal(t, expected, actual)
}

func TestGetRunningStepHeaderSubSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		ID:      longStr,
		Title:   longStr,
		Version: longStr,
	}

	actual := getRunningStepHeaderSubSection(stepInfo)
	require.NotEqual(t, "", actual)
}

func TestGetRunningStepFooterMainSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Title:   longStr,
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
	require.Equal(t, "| 🚫  | \x1b[31;1mThis is a very long string,\nthis is a very ... (exit code: 1)\x1b[0m| 0.01 sec |", cell)

	stepInfo.Title = ""
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		Error:    nil,
		ExitCode: 0,
	}

	cell = getRunningStepFooterMainSection(result)
	require.Equal(t, "| ✅  | \x1b[32;1m\x1b[0m                                                             | 0.00 sec |", cell)
}

func TestGetDeprecateNotesRows(t *testing.T) {
	notes := "Removal notes: " + longStr
	formattedNotes := getDeprecateNotesRows(notes)
	expected := "| \x1b[31;1mRemoval notes:\x1b[0m This is a very long string, this is a very long string, this  |\n| is a very long string, this is a very long string, this is a very long       |\n| string, this is a very long string.                                          |"
	require.Equal(t, expected, formattedNotes)
}

func TestGetRunningStepFooterSubSection(t *testing.T) {
	t.Log("Update available, no support_url, no source_code_url")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Title:   longStr,
			Version: "1.0.0",
			Latest:  "1.1.0",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
			Error:    errors.New(longStr),
			ExitCode: 1,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Update available: 1.0.0 -> 1.1.0                                             |\n| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |\n| Source: \x1b[33;1mNot provided\x1b[0m                                                         |"
		require.Equal(t, expected, actual)
	}

	t.Log("support url row length's chardiff = 0")
	{
		paddingCharCnt := 4
		placeholderCharCnt := len("Issue tracker: ")
		supportURLCharCnt := stepRunSummaryBoxWidthInChars - paddingCharCnt - placeholderCharCnt
		supportURL := strings.Repeat("a", supportURLCharCnt)

		// supportURL :=
		stepInfo := stepmanModels.StepInfoModel{
			Title:      longStr,
			Version:    "1.0.0",
			Latest:     "1.0.0",
			SupportURL: supportURL,
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
			Error:    errors.New(longStr),
			ExitCode: 1,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Issue tracker: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |\n| Source: \x1b[33;1mNot provided\x1b[0m                                                         |"
		require.Equal(t, expected, actual)
	}
}

func TestPrintRunningWorkflow(t *testing.T) {
	PrintRunningWorkflow(longStr)
}

func TestPrintRunningStepHeader(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Title:   "",
		Version: "",
	}
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.Title = longStr
	stepInfo.Version = ""
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.Title = ""
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, 0)

	stepInfo.Title = longStr
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, 0)
}

func TestPrintRunningStepFooter(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Title:   longStr,
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

	stepInfo.Title = ""
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
		Title:   longStr,
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

	stepInfo.Title = ""
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
