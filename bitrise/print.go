package bitrise

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/colorstring"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
)

const (
	stepRunSummaryBoxMaxWidthChars = 50

	// StepRunResultCodeSuccess ...
	StepRunResultCodeSuccess = 0
	// StepRunResultCodeFailed ...
	StepRunResultCodeFailed = 1
	// StepRunResultCodeFailedNotImportant ...
	StepRunResultCodeFailedNotImportant = 2
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
	content := fmt.Sprintf("(%d) %s", idx, title)
	if len(content) > stepRunSummaryBoxMaxWidthChars {
		dif := stepRunSummaryBoxMaxWidthChars - len(content)
		title = title[0 : len(content)-dif]
		content = fmt.Sprintf("(%d) %s", idx, title)
	}
	sep := strings.Repeat("-", len(content)+4)
	log.Info(sep)
	log.Infof("| (%d) %s |", idx, title)
	log.Info(sep)
}

// PrintStepSummary ..
func PrintStepSummary(title string, resultCode int, duration time.Duration, exitCode int) {
	runTime := TimeToFormattedSeconds(duration, " sec")
	content := fmt.Sprintf("%s | .... | %s", title, runTime)
	if resultCode == StepRunResultCodeFailed || resultCode == StepRunResultCodeFailedNotImportant {
		content = fmt.Sprintf("%s | .... | exit code: %d | %s", title, exitCode, runTime)
	}
	if len(content) > stepRunSummaryBoxMaxWidthChars {
		dif := len(content) - stepRunSummaryBoxMaxWidthChars
		title = title[0:(len(title) - dif)]
		content = fmt.Sprintf("%s | .... | %s", title, runTime)
		if resultCode == StepRunResultCodeFailed || resultCode == StepRunResultCodeFailedNotImportant {
			content = fmt.Sprintf("%s | .... | exit code: %d | %s", title, exitCode, runTime)
		}
	}

	sep := strings.Repeat("-", len(content)+2)
	log.Info(sep)
	switch resultCode {
	case StepRunResultCodeSuccess:
		runStateIcon := "✅ "
		content = fmt.Sprintf("%s | %s | %s", runStateIcon, colorstring.Green(title), runTime)
		break
	case StepRunResultCodeFailed:
		runStateIcon := "❌ "
		content = fmt.Sprintf("%s | %s | %s | exit code: %d", runStateIcon, colorstring.Red(title), runTime, exitCode)
		break
	case StepRunResultCodeFailedNotImportant:
		runStateIcon := "❌ "
		content = fmt.Sprintf("%s | %s | %s | exit code: %d", runStateIcon, colorstring.Yellow(title), runTime, exitCode)
		break
	case StepRunResultCodeSkipped, StepRunResultCodeSkippedWithRunIf:
		runStateIcon := "➡ "
		content = fmt.Sprintf("%s | %s | %s", runStateIcon, colorstring.White(title), runTime)
		break
	default:
		log.Error("Unkown result code")
		return
	}

	log.Infof("| " + content + " |")
	log.Info(sep)
	fmt.Println()
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
	failedNotImportantStepCount := 0
	skippedStepCount := 0

	successStepCount += len(buildRunResults.SuccessSteps)
	failedStepCount += len(buildRunResults.FailedSteps)
	failedNotImportantStepCount += len(buildRunResults.FailedNotImportantSteps)
	skippedStepCount += len(buildRunResults.SkippedSteps)
	totalStepCount = successStepCount + failedStepCount + failedNotImportantStepCount + skippedStepCount

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
		if failedNotImportantStepCount > 0 {
			log.Info(colorstring.Yellowf(" * %d failed but was marked as skippable and", failedNotImportantStepCount))
		}
		if skippedStepCount > 0 {
			log.Info(colorstring.Whitef(" * %d was skipped", skippedStepCount))
		}

		fmt.Println()
		if failedStepCount > 0 {
			log.Fatal("FINISHED but a couple of steps failed - Ouch")
		} else {
			log.Info("DONE - Congrats!!")
			if failedNotImportantStepCount > 0 {
				log.Warn("P.S.: a couple of non imporatant steps failed")
			}
		}
	}
}

// PrintStepStatus ...
func PrintStepStatus(stepRunResults models.BuildRunResultsModel) {
	failedCount := len(stepRunResults.FailedSteps)
	failedNotImportantCount := len(stepRunResults.FailedNotImportantSteps)
	skippedCount := len(stepRunResults.SkippedSteps)
	successCount := len(stepRunResults.SuccessSteps)
	totalCount := successCount + failedCount + failedNotImportantCount + skippedCount

	log.Infof("Out of %d steps, %d was successful, %d failed, %d failed but was marked as skippable and %d was skipped",
		totalCount,
		successCount,
		failedCount,
		failedNotImportantCount,
		skippedCount)

	PrintStepStatusList("Failed steps:", stepRunResults.FailedSteps)
	PrintStepStatusList("Failed but skippable steps:", stepRunResults.FailedNotImportantSteps)
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
