package bitrise

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 65
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
		trimmedTitleWidth := len(title) - charDiff - 3
		content = fmt.Sprintf("| (%d) %s... |", idx, title[0:trimmedTitleWidth])
	}

	sep := strings.Repeat("-", len(content))
	log.Info(sep)
	log.Infof(content)
	log.Info(sep)
	log.Info("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")
}

func getTrimmedStepName(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	title := ""
	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess:
		title = stepRunResult.StepName
		if len(title) > titleBoxWidth {
			dif := len(title) - titleBoxWidth
			title = title[:len(title)-dif-3] + "..."
		}
		break
	case models.StepRunStatusCodeFailed:
		title = fmt.Sprintf("%s (exit code: %d)", stepRunResult.StepName, stepRunResult.ExitCode)
		if len(title) > titleBoxWidth {
			dif := len(title) - titleBoxWidth
			title = title[:len(stepRunResult.StepName)-dif-3] + "..."
			title = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		}
		break
	case models.StepRunStatusCodeFailedSkippable:
		title = fmt.Sprintf("%s (exit code: %d)", stepRunResult.StepName, stepRunResult.ExitCode)
		if len(title) > titleBoxWidth {
			dif := len(title) - titleBoxWidth
			title = title[:len(stepRunResult.StepName)-dif-3] + "..."
			title = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		}
		break
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		title = stepRunResult.StepName
		if len(title) > titleBoxWidth {
			dif := len(title) - titleBoxWidth
			title = title[:len(title)-dif-3] + "..."
		}
		break
	default:
		log.Error("Unkown result code")
		return ""
	}
	return title
}

func stepResultCell(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	icon := ""
	title := getTrimmedStepName(stepRunResult)
	runTimeStr := TimeToFormattedSeconds(stepRunResult.RunTime, " sec")
	coloringFunc := colorstring.Green
	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess:
		icon = "✅"
		coloringFunc = colorstring.Green
		break
	case models.StepRunStatusCodeFailed:
		icon = "🚫"
		coloringFunc = colorstring.Red
		break
	case models.StepRunStatusCodeFailedSkippable:
		icon = "⚠️"
		coloringFunc = colorstring.Yellow
		break
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		icon = "➡"
		coloringFunc = colorstring.Blue
		break
	default:
		log.Error("Unkown result code")
		return ""
	}

	iconBox := fmt.Sprintf(" %s  ", icon)

	titleWhiteSpaceWidth := titleBoxWidth - len(title)
	titleBox := fmt.Sprintf(" %s%s", coloringFunc(title), strings.Repeat(" ", titleWhiteSpaceWidth))

	timeWhiteSpaceWidth := timeBoxWidth - len(runTimeStr) - 1
	timeBox := fmt.Sprintf(" %s%s", runTimeStr, strings.Repeat(" ", timeWhiteSpaceWidth))

	return fmt.Sprintf("|%s|%s|%s|", iconBox, titleBox, timeBox)
}

// PrintStepSummary ..
func PrintStepSummary(stepRunResult models.StepRunResultsModel, isLastStepInWorkflow bool) {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth
	sep := fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	log.Info("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")

	log.Info(sep)
	log.Infof(stepResultCell(stepRunResult))
	log.Info(sep)

	if !isLastStepInWorkflow {
		fmt.Println()
		fmt.Println(strings.Repeat(" ", 42) + "▼")
		fmt.Println()
	}
}

// PrintSummary ...
func PrintSummary(buildRunResults models.BuildRunResultsModel) {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth

	fmt.Println()
	fmt.Println()
	log.Infof("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))
	whitespaceWidth := (stepRunSummaryBoxWidthInChars - 2 - len("bitrise summary")) / 2
	log.Infof("|%sbitrise summary%s|", strings.Repeat(" ", whitespaceWidth), strings.Repeat(" ", whitespaceWidth))
	log.Infof("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	whitespaceWidth = stepRunSummaryBoxWidthInChars - len("|    | title") - len("| time (s) |")
	log.Infof("|    | title%s| time (s) |", strings.Repeat(" ", whitespaceWidth))
	log.Infof("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	orderedResults := buildRunResults.OrderedResults()
	for _, stepRunResult := range orderedResults {
		log.Info(stepResultCell(stepRunResult))
	}

	log.Infof("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))
	fmt.Println()
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
