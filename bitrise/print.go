package bitrise

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	"github.com/bitrise-io/go-utils/colorstring"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 65

	// StepRunResultCodeSuccess ...
	StepRunResultCodeSuccess = 0
	// StepRunResultCodeFailed ...
	StepRunResultCodeFailed = 1
	// StepRunResultCodeFailedSkippable ...
	StepRunResultCodeFailedSkippable = 2
	// StepRunResultCodeSkipped ...
	StepRunResultCodeSkipped = 3
	// StepRunResultCodeSkippedWithRunIf ...
	StepRunResultCodeSkippedWithRunIf = 4
)

// PrintRunningWorkflow ...
func PrintRunningWorkflow(title string) {
	fmt.Println()
	log.Info(colorstring.Bluef("Running workflow (%s)", title))
	fmt.Println()
}

// PrintRunningStep ...
func PrintRunningStep(title string, idx int) {
	content := fmt.Sprintf("| (%d) %s |", idx, title)
	charDiff := len(content) - stepRunSummaryBoxWidthInChars

	if charDiff < 0 {
		// shorter than desired - fill with space
		content = fmt.Sprintf("| (%d) %s%s |", idx, title, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedTitleLength := len(title) - charDiff - 3
		content = fmt.Sprintf("| (%d) %s... |", idx, title[0:trimmedTitleLength])
	}

	sep := strings.Repeat("-", len(content))
	log.Info(sep)
	log.Infof(content)
	log.Info(sep)
	log.Info("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")
	// fmt.Println()
}

//
type coloringFn func(string) string

func stepSummaryString(runStateIcon, stepTitle, runTimeString string, stepExitCode int, coloringFunc coloringFn) string {
	contentLayout := ""
	if stepExitCode == 0 {
		contentLayout = fmt.Sprintf("| %s | %s | %s |", runStateIcon, stepTitle, runTimeString)
	} else {
		contentLayout = fmt.Sprintf("| %s | %s | %s | exit code: %d |", runStateIcon, stepTitle, runTimeString, stepExitCode)
	}
	content := ""
	charDiff := len(contentLayout) - stepRunSummaryBoxWidthInChars
	if charDiff < 0 {
		// shorter than desired - fill with space
		if stepExitCode == 0 {
			content = fmt.Sprintf("| %s | %s%s | %s |", runStateIcon, coloringFunc(stepTitle), strings.Repeat(" ", -charDiff), runTimeString)
		} else {
			content = fmt.Sprintf("| %s | %s%s | %s | exit code: %d |", runStateIcon, coloringFunc(stepTitle), strings.Repeat(" ", -charDiff), runTimeString, stepExitCode)
		}
		return content
	} else if charDiff > 0 {
		// longer than desired - trim stepTitle
		trimmedTitleLength := len(stepTitle) - charDiff - 3
		fmt.Println("trimmedTitleLength: ", trimmedTitleLength)
		if stepExitCode == 0 {
			content = fmt.Sprintf("| %s | %s... | %s |", runStateIcon, coloringFunc(stepTitle[0:trimmedTitleLength]), runTimeString)
		} else {
			content = fmt.Sprintf("| %s | %s... | %s | exit code: %d |", runStateIcon, coloringFunc(stepTitle[0:trimmedTitleLength]), runTimeString, stepExitCode)
		}
		return content
	}

	if stepExitCode == 0 {
		content = fmt.Sprintf("| %s | %s | %s |", runStateIcon, coloringFunc(stepTitle), runTimeString)
	} else {
		content = fmt.Sprintf("| %s | %s | %s | exit code: %d |", runStateIcon, coloringFunc(stepTitle), runTimeString, stepExitCode)
	}
	return content
}

// PrintStepSummary ..
func PrintStepSummary(title string, resultCode int, duration time.Duration, exitCode int, isLastStepInWorkflow bool) {
	runTimeStr := TimeToFormattedSeconds(duration, " sec")

	content := ""
	switch resultCode {
	case StepRunResultCodeSuccess:
		content = stepSummaryString("✅ ", title, runTimeStr, 0, colorstring.Green)
		break
	case StepRunResultCodeFailed:
		content = stepSummaryString("🚫 ", title, runTimeStr, exitCode, colorstring.Red)
		break
	case StepRunResultCodeFailedSkippable:
		content = stepSummaryString("⚠️ ", title, runTimeStr, exitCode, colorstring.Yellow)
		break
	case StepRunResultCodeSkipped, StepRunResultCodeSkippedWithRunIf:
		content = stepSummaryString("➡", title, runTimeStr, 0, colorstring.Blue)
		break
	default:
		log.Error("Unkown result code")
		return
	}

	sep := strings.Repeat("-", stepRunSummaryBoxWidthInChars)

	// fmt.Println()
	log.Info("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")
	log.Info(sep)
	log.Infof(content)
	log.Info(sep)

	if !isLastStepInWorkflow {
		fmt.Println()
		fmt.Println(strings.Repeat(" ", 42) + "▼")
		fmt.Println()
	}
}

// PrintBuildFailedFatal ...
func PrintBuildFailedFatal(startTime time.Time, err error) {
	runTime := time.Now().Sub(startTime)
	log.Error("Build failed: " + err.Error())
	log.Fatal("Total run time: " + runTime.String())
}

// PrintSummary ...
func PrintSummary(buildRunResults models.BuildRunResultsModel) {
	totalStepCount := 0
	successStepCount := 0
	failedStepCount := 0
	failedSkippableStepCount := 0
	skippedStepCount := 0

	successStepCount += len(buildRunResults.SuccessSteps)
	failedStepCount += len(buildRunResults.FailedSteps)
	failedSkippableStepCount += len(buildRunResults.FailedSkippableSteps)
	skippedStepCount += len(buildRunResults.SkippedSteps)
	totalStepCount = successStepCount + failedStepCount + failedSkippableStepCount + skippedStepCount

	fmt.Println()
	log.Infoln("==> Summary:")
	runTime := time.Now().Sub(buildRunResults.StartTime)
	log.Info("Total run time: " + TimeToFormattedSeconds(runTime, " seconds"))

	if totalStepCount > 0 {
		log.Infof("Out of %d steps:", totalStepCount)

		if successStepCount > 0 {
			log.Info(colorstring.Greenf(" * %d was successful", successStepCount))
		}
		if failedStepCount > 0 {
			log.Info(colorstring.Redf(" * %d failed", failedStepCount))
		}
		if failedSkippableStepCount > 0 {
			log.Info(colorstring.Yellowf(" * %d failed but was marked as skippable and", failedSkippableStepCount))
		}
		if skippedStepCount > 0 {
			log.Info(colorstring.Bluef(" * %d was skipped", skippedStepCount))
		}

		fmt.Println()
		if failedStepCount > 0 {
			log.Error("FINISHED but a couple of steps failed - Ouch")
		} else {
			log.Info("DONE - Congrats!!")
			if failedSkippableStepCount > 0 {
				log.Warn("P.S.: a couple of non imporatant steps failed")
			}
		}
	}
}

// PrintStepStatus ...
func PrintStepStatus(stepRunResults models.BuildRunResultsModel) {
	failedCount := len(stepRunResults.FailedSteps)
	failedSkippableCount := len(stepRunResults.FailedSkippableSteps)
	skippedCount := len(stepRunResults.SkippedSteps)
	successCount := len(stepRunResults.SuccessSteps)
	totalCount := successCount + failedCount + failedSkippableCount + skippedCount

	log.Infof("Out of %d steps, %d was successful, %d failed, %d failed but was marked as skippable and %d was skipped",
		totalCount,
		successCount,
		failedCount,
		failedSkippableCount,
		skippedCount)

	PrintStepStatusList("Failed steps:", stepRunResults.FailedSteps)
	PrintStepStatusList("Failed but skippable steps:", stepRunResults.FailedSkippableSteps)
	PrintStepStatusList("Skipped steps:", stepRunResults.SkippedSteps)
}

// PrintStepStatusList ...
func PrintStepStatusList(header string, stepList []models.StepRunResultsModel) {
	if len(stepList) > 0 {
		log.Infof(header)
		for _, step := range stepList {
			if step.Error != nil {
				log.Infof(" * Step: (%s) | error: (%v)", step.StepName, step.Error)
			} else {
				log.Infof(" * Step: (%s)", step.StepName)
			}
		}
	}
}
