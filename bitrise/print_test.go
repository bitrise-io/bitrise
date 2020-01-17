package bitrise

import (
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

const longStr = "This is a very long string, this is a very long string, " +
	"this is a very long string, this is a very long string," +
	"this is a very long string, this is a very long string."

func TestIsUpdateAvailable(t *testing.T) {
	t.Log("simple compare versions - ture")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "1.1.0",
		}

		require.Equal(t, true, isUpdateAvailable(stepInfo1))
	}

	t.Log("simple compare versions - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "1.0.0",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}

	t.Log("issue - no latest - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}

	t.Log("issue - no current - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "",
			LatestVersion: "1.0.0",
		}

		require.Equal(t, false, isUpdateAvailable(stepInfo1))
	}
}

func TestGetTrimmedStepName(t *testing.T) {
	t.Log("successful step")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(longStr),
			},
			Version: longStr,
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
			ErrorStr: longStr,
			ExitCode: 1,
		}

		actual := getTrimmedStepName(result)
		expected := "This is a very long string, this is a very long string, thi..."
		require.Equal(t, expected, actual)
	}

	t.Log("failed step")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(""),
			},
			Version: longStr,
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  0,
			ErrorStr: "",
			ExitCode: 0,
		}

		actual := getTrimmedStepName(result)
		expected := ""
		require.Equal(t, expected, actual)
	}
}

func TestGetRunningStepHeaderMainSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Step: stepmanModels.StepModel{
			Title: pointers.NewStringPtr(longStr),
		},
		Version: longStr,
	}

	actual := getRunningStepHeaderMainSection(stepInfo, 0)
	expected := "| (0) This is a very long string, this is a very long string, this is a ver... |"
	require.Equal(t, expected, actual)
}

func TestGetRunningStepHeaderSubSection(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		ID: longStr,
		Step: stepmanModels.StepModel{
			Title: pointers.NewStringPtr(longStr),
		},
		Version: longStr,
	}

	actual := getRunningStepHeaderSubSection(stepmanModels.StepModel{}, stepInfo)
	require.NotEqual(t, "", actual)
}

func TestGetRunningStepFooterMainSection(t *testing.T) {
	t.Log("failed step")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(longStr),
			},
			Version: longStr,
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeFailed,
			Idx:      0,
			RunTime:  10000000,
			ErrorStr: longStr,
			ExitCode: 1,
		}

		actual := getRunningStepFooterMainSection(result)
		expected := "| \x1b[31;1mx\x1b[0m | \x1b[31;1mThis is a very long string, this is a very l... (exit code: 1)\x1b[0m| 0.01 sec |"
		require.Equal(t, expected, actual)
	}

	t.Log("successful step")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(""),
			},
			Version: longStr,
		}
		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  0,
			ErrorStr: "",
			ExitCode: 0,
		}

		actual := getRunningStepFooterMainSection(result)
		expected := "| \x1b[32;1m✓\x1b[0m | \x1b[32;1m\x1b[0m                                                              | 0.00 sec |"
		require.Equal(t, expected, actual)
	}

	t.Log("long Runtime")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(""),
			},
			Version: longStr,
		}
		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  100 * 1000 * 1e9, // 100 * 1000 * 10^9 nanosec = 100 000 sec
			ErrorStr: "",
			ExitCode: 0,
		}

		actual := getRunningStepFooterMainSection(result)
		expected := "| \x1b[32;1m✓\x1b[0m | \x1b[32;1m\x1b[0m                                                              | 28 hour  |"
		require.Equal(t, expected, actual)
	}

	t.Log("long Runtime")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(""),
			},
			Version: longStr,
		}
		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  hourToDuration(1000),
			ErrorStr: "",
			ExitCode: 0,
		}

		actual := getRunningStepFooterMainSection(result)
		expected := "| \x1b[32;1m✓\x1b[0m | \x1b[32;1m\x1b[0m                                                              | 999+ hour|"
		require.Equal(t, expected, actual)
	}
}

func TestGetDeprecateNotesRows(t *testing.T) {
	notes := "Removal notes: " + longStr
	actual := getDeprecateNotesRows(notes)
	expected := "| \x1b[31;1mRemoval notes:\x1b[0m This is a very long string, this is a very long string, this  |" + "\n" +
		"| is a very long string, this is a very long string,this is a very long        |" + "\n" +
		"| string, this is a very long string.                                          |"
	require.Equal(t, expected, actual)
}

func TestGetRunningStepFooterSubSection(t *testing.T) {
	t.Log("Update available, no support_url, no source_code_url")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title: pointers.NewStringPtr(longStr),
			},
			Version:          "1.0.0",
			LatestVersion:    "1.1.0",
			EvaluatedVersion: "1.0.0",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
			ErrorStr: longStr,
			ExitCode: 1,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Update available: 1.0.0 -> 1.1.0                                             |" + "\n" +
			"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |" + "\n" +
			"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |"
		require.Equal(t, expected, actual)
	}

	t.Log("Update available, major/minor lock, with changelog URL cropping")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title:         pointers.NewStringPtr(longStr),
				SourceCodeURL: pointers.NewStringPtr("https://github.com/test-organization/very-long-test-repository-name-exceeding-max-width"),
			},
			Version:          "1",
			LatestVersion:    "2.1.0",
			EvaluatedVersion: "1.0.1",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Update available: 1 (1.0.1) -> 2.1.0                                         |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| ...-organization/very-long-test-repository-name-exceeding-max-width/releases |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.x.x"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.x.x (1.0.1) -> 2.1.0                                     |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| ...-organization/very-long-test-repository-name-exceeding-max-width/releases |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.0"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.0 (1.0.1) -> 2.1.0                                       |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| ...-organization/very-long-test-repository-name-exceeding-max-width/releases |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.0.x"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.0.x (1.0.1) -> 2.1.0                                     |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| ...-organization/very-long-test-repository-name-exceeding-max-width/releases |"
		require.Equal(t, expected, actual)

	}

	t.Log("Update available, major/minor lock, without changelog URL cropping")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title:         pointers.NewStringPtr(longStr),
				SourceCodeURL: pointers.NewStringPtr("https://github.com/bitrise-steplib/steps-script"),
			},
			Version:          "1",
			LatestVersion:    "2.1.0",
			EvaluatedVersion: "1.0.1",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Update available: 1 (1.0.1) -> 2.1.0                                         |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| https://github.com/bitrise-steplib/steps-script/releases                     |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.x.x"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.x.x (1.0.1) -> 2.1.0                                     |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| https://github.com/bitrise-steplib/steps-script/releases                     |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.0"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.0 (1.0.1) -> 2.1.0                                       |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| https://github.com/bitrise-steplib/steps-script/releases                     |"
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.0.x"
		actual = getRunningStepFooterSubSection(result)
		expected = "| Update available: 1.0.x (1.0.1) -> 2.1.0                                     |" + "\n" +
			"|                                                                              |" + "\n" +
			"| Release notes are available on GitHub                                        |" + "\n" +
			"| https://github.com/bitrise-steplib/steps-script/releases                     |"
		require.Equal(t, expected, actual)

	}

	t.Log("Update available, nothing is printed if latest version is within major/minor lock range")
	{
		stepInfo := stepmanModels.StepInfoModel{
			Step: stepmanModels.StepModel{
				Title:         pointers.NewStringPtr(longStr),
				SourceCodeURL: pointers.NewStringPtr("https://github.com/bitrise-steplib/steps-script"),
			},
			Version:          "1",
			LatestVersion:    "1.0.1",
			EvaluatedVersion: "1.0.1",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := ""
		require.Equal(t, expected, actual)

		result.StepInfo.Version = "1.0"
		actual = getRunningStepFooterSubSection(result)
		expected = ""
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
			Step: stepmanModels.StepModel{
				Title:      pointers.NewStringPtr(longStr),
				SupportURL: pointers.NewStringPtr(supportURL),
			},
			Version:       "1.0.0",
			LatestVersion: "1.0.0",
		}

		result := models.StepRunResultsModel{
			StepInfo: stepInfo,
			Status:   models.StepRunStatusCodeSuccess,
			Idx:      0,
			RunTime:  10000000,
			ErrorStr: longStr,
			ExitCode: 1,
		}

		actual := getRunningStepFooterSubSection(result)
		expected := "| Issue tracker: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |" + "\n" +
			"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |"
		require.Equal(t, expected, actual)
	}
}

func TestPrintRunningWorkflow(t *testing.T) {
	PrintRunningWorkflow(longStr)
}

func TestPrintRunningStepHeader(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Step: stepmanModels.StepModel{
			Title: pointers.NewStringPtr(""),
		},
		Version: "",
	}
	step := stepmanModels.StepModel{}
	PrintRunningStepHeader(stepInfo, step, 0)

	stepInfo.Step.Title = pointers.NewStringPtr(longStr)
	stepInfo.Version = ""
	PrintRunningStepHeader(stepInfo, step, 0)

	stepInfo.Step.Title = pointers.NewStringPtr("")
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, step, 0)

	stepInfo.Step.Title = pointers.NewStringPtr(longStr)
	stepInfo.Version = longStr
	PrintRunningStepHeader(stepInfo, step, 0)
}

func TestPrintRunningStepFooter(t *testing.T) {
	stepInfo := stepmanModels.StepInfoModel{
		Step: stepmanModels.StepModel{
			Title: pointers.NewStringPtr(longStr),
		},
		Version: longStr,
	}

	result := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		ErrorStr: longStr,
		ExitCode: 1,
	}
	PrintRunningStepFooter(result, true)
	PrintRunningStepFooter(result, false)

	stepInfo.Step.Title = pointers.NewStringPtr("")
	result = models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		ErrorStr: "",
		ExitCode: 0,
	}
	PrintRunningStepFooter(result, true)
	PrintRunningStepFooter(result, false)
}

func TestPrintSummary(t *testing.T) {
	PrintSummary(models.BuildRunResultsModel{})

	stepInfo := stepmanModels.StepInfoModel{
		Step: stepmanModels.StepModel{
			Title: pointers.NewStringPtr(longStr),
		},
		Version: longStr,
	}

	result1 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  10000000,
		ErrorStr: longStr,
		ExitCode: 1,
	}

	stepInfo.Step.Title = pointers.NewStringPtr("")
	result2 := models.StepRunResultsModel{
		StepInfo: stepInfo,
		Status:   models.StepRunStatusCodeSuccess,
		Idx:      0,
		RunTime:  0,
		ErrorStr: "",
		ExitCode: 0,
	}

	buildResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
		SuccessSteps:   []models.StepRunResultsModel{result1, result2},
	}

	PrintSummary(buildResults)
}
