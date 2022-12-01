package bitrise

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/stringutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 80
)

//------------------------------
// Util methods
//------------------------------

func trimTitle(title string, titleSuffix string, titleBoxWidth int) string {
	length := len(title)
	if titleSuffix != "" {
		length += 1 + len(titleSuffix)
	}

	if length > titleBoxWidth {
		diff := length - titleBoxWidth
		title = stringutil.MaxFirstCharsWithDots(title, len(title)-diff)
	}

	if titleSuffix == "" {
		return title
	}

	return fmt.Sprintf("%s %s", title, titleSuffix)
}

func getTrimmedStepName(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("   ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	stepInfo := stepRunResult.StepInfo

	title := ""
	if stepInfo.Step.Title != nil && *stepInfo.Step.Title != "" {
		title = *stepInfo.Step.Title
	}

	if stepInfo.GroupInfo.RemovalDate != "" {
		title = fmt.Sprintf("[Deprecated] %s", title)
	}

	suffix := ""
	reason := stepRunResult.StatusReason()
	if reason != "" {
		suffix = fmt.Sprintf("(%s)", reason)
	}

	return trimTitle(title, suffix, titleBoxWidth)
}

func getRunningStepFooterMainSection(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("   ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	icon := ""
	title := getTrimmedStepName(stepRunResult)
	coloringFunc := colorstring.Green

	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess:
		icon = "✓"
		coloringFunc = colorstring.Green
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodePreparationFailed:
		icon = "x"
		coloringFunc = colorstring.Red
	case models.StepRunStatusAbortedWithCustomTimeout, models.StepRunStatusAbortedWithNoOutputTimeout:
		icon = "/"
		coloringFunc = colorstring.Red
	case models.StepRunStatusCodeFailedSkippable:
		icon = "!"
		coloringFunc = colorstring.Yellow
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		icon = "-"
		coloringFunc = colorstring.Blue
	default:
		log.Errorf("Unknown result code")
		return ""
	}

	iconBox := fmt.Sprintf(" %s ", coloringFunc(icon))

	titleWhiteSpaceWidth := titleBoxWidth - len(title)
	coloredTitle := title
	if strings.HasPrefix(title, "[Deprecated]") {
		title := strings.TrimPrefix(title, "[Deprecated]")
		coloredTitle = fmt.Sprintf("%s%s", colorstring.Red("[Deprecated]"), coloringFunc(title))
	} else {
		coloredTitle = coloringFunc(title)
	}

	titleBox := fmt.Sprintf(" %s%s", coloredTitle, strings.Repeat(" ", titleWhiteSpaceWidth))

	runTimeStr, err := utils.FormattedSecondsToMax8Chars(stepRunResult.RunTime)
	if err != nil {
		log.Errorf("Failed to format time, error: %s", err)
		runTimeStr = "999+ hour"
	}

	timeWhiteSpaceWidth := timeBoxWidth - len(runTimeStr) - 1
	if timeWhiteSpaceWidth < 0 {
		log.Errorf("Invalid time box size for RunTime: %#v", stepRunResult.RunTime)
		timeWhiteSpaceWidth = 0
	}
	timeBox := fmt.Sprintf(" %s%s", runTimeStr, strings.Repeat(" ", timeWhiteSpaceWidth))

	return fmt.Sprintf("|%s|%s|%s|", iconBox, titleBox, timeBox)
}

func getDeprecateNotesRows(notes string) string {
	colorDeprecateNote := func(line string) string {
		if strings.HasPrefix(line, "Removal notes:") {
			line = strings.TrimPrefix(line, "Removal notes:")
			line = fmt.Sprintf("%s%s", colorstring.Red("Removal notes:"), line)
		}
		return line
	}

	boxContentWidth := stepRunSummaryBoxWidthInChars - 4

	notesWithoutNewLine := strings.Replace(notes, "\n", " ", -1)
	words := strings.Split(notesWithoutNewLine, " ")
	if len(words) == 0 {
		return ""
	}

	formattedNote := ""
	line := ""

	for i, word := range words {
		isLastLine := i == len(words)-1

		expectedLine := ""
		if line == "" {
			expectedLine = word
		} else {
			expectedLine = line + " " + word
		}

		if utf8.RuneCountInString(expectedLine) > boxContentWidth {
			// expected line would be to long, so print the previous line, and start a new with the last word.
			noteRow := fmt.Sprintf("| %s |", line)
			charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
			if charDiff <= 0 {
				// shorter than desired - fill with space
				line = colorDeprecateNote(line)
				noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
			} else if charDiff > 0 {
				// longer than desired - should not
				log.Errorf("Should not be longer then expected")
			}

			if formattedNote == "" {
				formattedNote = noteRow
			} else {
				formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
			}

			line = word

			if isLastLine {
				noteRow := fmt.Sprintf("| %s |", line)
				charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
				if charDiff < 0 {
					// shorter than desired - fill with space
					line = colorDeprecateNote(line)
					noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
				} else if charDiff > 0 {
					// longer than desired - should not
					log.Errorf("Should not be longer then expected")
				}

				if formattedNote == "" {
					formattedNote = noteRow
				} else {
					formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
				}
			}
		} else {
			// expected line is not to long, just keep growing the line
			line = expectedLine

			if isLastLine {
				noteRow := fmt.Sprintf("| %s |", line)
				charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
				if charDiff <= 0 {
					// shorter than desired - fill with space
					line = colorDeprecateNote(line)
					noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
				} else if charDiff > 0 {
					// longer than desired - should not
					log.Errorf("Should not be longer then expected")
				}

				if formattedNote == "" {
					formattedNote = noteRow
				} else {
					formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
				}
			}
		}
	}

	return formattedNote
}

func getRow(str string) string {
	str = stringutil.MaxLastCharsWithDots(str, stepRunSummaryBoxWidthInChars-4)
	return fmt.Sprintf("| %s |", str+strings.Repeat(" ", stepRunSummaryBoxWidthInChars-len(str)-4))
}

func getUpdateRow(stepInfo stepmanModels.StepInfoModel, width int) string {
	vstr := fmt.Sprintf("%s -> %s", stepInfo.Version, stepInfo.LatestVersion)
	if stepInfo.Version != stepInfo.OriginalVersion {
		vstr = fmt.Sprintf("%s (%s) -> %s", stepInfo.OriginalVersion, stepInfo.Version, stepInfo.LatestVersion)
	}

	updateRow := fmt.Sprintf("| Update available: %s |", vstr)
	charDiff := len(updateRow) - width

	if charDiff == 0 {
		return updateRow
	}

	// shorter than desired - fill with space
	updateRow = fmt.Sprintf("| Update available: %s%s |", vstr, strings.Repeat(" ", -charDiff))

	if charDiff > 0 {
		// longer than desired - trim title
		updateRow = fmt.Sprintf("| Update available: -> %s%s |", stepInfo.LatestVersion, strings.Repeat(" ", -len("| Update available: -> %s |")-width))
		if charDiff > 6 {
			updateRow = fmt.Sprintf("| Update available!%s |", strings.Repeat(" ", -len("| Update available! |")-width))
		}
	}

	return updateRow
}

func getRunningStepFooterSubSection(stepRunResult models.StepRunResultsModel) string {
	stepInfo := stepRunResult.StepInfo

	removalDate := stepInfo.GroupInfo.RemovalDate
	deprecateNotes := stepInfo.GroupInfo.DeprecateNotes
	removalDateRow := ""
	deprecateNotesRow := ""
	if removalDate != "" {
		removalDateValue := removalDate
		removalDateKey := colorstring.Red("Removal date:")

		removalDateRow = fmt.Sprintf("| Removal date: %s |", removalDateValue)
		charDiff := len(removalDateRow) - stepRunSummaryBoxWidthInChars
		removalDateRow = fmt.Sprintf("| %s %s%s |", removalDateKey, removalDateValue, strings.Repeat(" ", -charDiff))

		if deprecateNotes != "" {
			deprecateNotesStr := fmt.Sprintf("Removal notes: %s", deprecateNotes)
			deprecateNotesRow = getDeprecateNotesRows(deprecateNotesStr)
		}
	}

	isUpdateAvailable, err := utils.IsUpdateAvailable(stepRunResult.StepInfo.Version, stepRunResult.StepInfo.LatestVersion)
	if err != nil {
		log.Warn(err)
	}

	updateRow := ""
	if isUpdateAvailable {
		updateRow = getUpdateRow(stepInfo, stepRunSummaryBoxWidthInChars)
	}

	issueRow := ""
	sourceRow := ""
	if stepRunResult.ErrorStr != "" {
		// Support URL
		var coloringFunc func(...interface{}) string
		supportURL := ""
		if stepInfo.Step.SupportURL != nil && *stepInfo.Step.SupportURL != "" {
			supportURL = *stepInfo.Step.SupportURL
		}
		if supportURL == "" {
			coloringFunc = colorstring.Yellow
			supportURL = "Not provided"
		}

		issueRow = fmt.Sprintf("| Issue tracker: %s |", supportURL)

		charDiff := len(issueRow) - stepRunSummaryBoxWidthInChars
		if charDiff <= 0 {
			// shorter than desired - fill with space

			if coloringFunc != nil {
				// We need to do this after charDiff calculation,
				// because of coloring characters increase the text length, but they do not printed
				supportURL = coloringFunc("Not provided")
			}

			issueRow = fmt.Sprintf("| Issue tracker: %s%s |", supportURL, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(supportURL) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Support url too long, can't present support url at all! : %s", supportURL)
			} else {
				issueRow = fmt.Sprintf("| Issue tracker: %s |", stringutil.MaxLastCharsWithDots(supportURL, trimmedWidth))
			}
		}

		// Source Code URL
		coloringFunc = nil
		sourceCodeURL := ""
		if stepInfo.Step.SourceCodeURL != nil && *stepInfo.Step.SourceCodeURL != "" {
			sourceCodeURL = *stepInfo.Step.SourceCodeURL
		}
		if sourceCodeURL == "" {
			coloringFunc = colorstring.Yellow
			sourceCodeURL = "Not provided"
		}

		sourceRow = fmt.Sprintf("| Source: %s |", sourceCodeURL)

		charDiff = len(sourceRow) - stepRunSummaryBoxWidthInChars
		if charDiff <= 0 {
			// shorter than desired - fill with space

			if coloringFunc != nil {
				// We need to do this after charDiff calculation,
				// because of coloring characters increase the text length, but they do not printed
				sourceCodeURL = coloringFunc("Not provided")
			}

			sourceRow = fmt.Sprintf("| Source: %s%s |", sourceCodeURL, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(sourceCodeURL) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Source url too long, can't present source url at all! : %s", sourceCodeURL)
			} else {
				sourceRow = fmt.Sprintf("| Source: %s |", stringutil.MaxLastCharsWithDots(sourceCodeURL, trimmedWidth))
			}
		}
	}

	// Update available
	content := ""
	if isUpdateAvailable {
		content = updateRow
		if stepInfo.Step.SourceCodeURL != nil && *stepInfo.Step.SourceCodeURL != "" {
			content += "\n" + getRow("")
			releasesURL := utils.RepoReleasesURL(*stepInfo.Step.SourceCodeURL)
			content += "\n" + getRow("Release notes are available below")
			content += "\n" + getRow(releasesURL)
		}
	}

	// Support URL
	if issueRow != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, issueRow)
		} else {
			content = fmt.Sprintf("%s", issueRow)
		}
	}

	// Source Code URL
	if sourceRow != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, sourceRow)
		} else {
			content = fmt.Sprintf("%s", sourceRow)
		}
	}

	// Deprecation
	if removalDate != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, removalDateRow)
		} else {
			content = fmt.Sprintf("%s", removalDateRow)
		}

		if deprecateNotes != "" {
			if content != "" {
				content = fmt.Sprintf("%s\n%s", content, deprecateNotesRow)
			} else {
				content = fmt.Sprintf("%s", deprecateNotesRow)
			}
		}
	}

	return content
}

// PrintRunningWorkflow ...
func PrintRunningWorkflow(title string) {
	log.Print()
	log.Infof("Switching to workflow: %s", title)
}

// PrintSummary ...
func PrintSummary(buildRunResults models.BuildRunResultsModel) {
	iconBoxWidth := len("   ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth

	log.Print()
	log.Print()
	log.Printf("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	title := fmt.Sprintf("bitrise summary: %s", buildRunResults.WorkflowID)
	whitespace := float64(stepRunSummaryBoxWidthInChars - 2 - len(title))
	if whitespace < 0 {
		whitespace = 0
	}
	leftPadding := int(math.Floor(whitespace / 2.0))
	rightPadding := int(math.Ceil(whitespace / 2.0))
	log.Printf("|%s%s%s|", strings.Repeat(" ", leftPadding), title, strings.Repeat(" ", rightPadding))
	log.Printf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	whitespaceWidth := stepRunSummaryBoxWidthInChars - len("|   | title") - len("| time (s) |")
	log.Printf("|   | title%s| time (s) |", strings.Repeat(" ", whitespaceWidth))
	log.Printf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	orderedResults := buildRunResults.OrderedResults()
	tmpTime := time.Time{}
	for _, stepRunResult := range orderedResults {
		tmpTime = tmpTime.Add(stepRunResult.RunTime)
		log.Print(getRunningStepFooterMainSection(stepRunResult))
		log.Printf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

		updateAvailable, _ := utils.IsUpdateAvailable(stepRunResult.StepInfo.Version, stepRunResult.StepInfo.LatestVersion)

		if stepRunResult.ErrorStr != "" || stepRunResult.StepInfo.GroupInfo.RemovalDate != "" || updateAvailable {
			footerSubSection := getRunningStepFooterSubSection(stepRunResult)
			if footerSubSection != "" {
				log.Print(footerSubSection)
				log.Printf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))
			}
		}
	}
	runtime := tmpTime.Sub(time.Time{})

	runTimeStr, err := utils.FormattedSecondsToMax8Chars(runtime)
	if err != nil {
		log.Errorf("Failed to format time, error: %s", err)
		runTimeStr = "999+ hour"
	}

	whitespaceWidth = stepRunSummaryBoxWidthInChars - len(fmt.Sprintf("| Total runtime: %s|", runTimeStr))
	if whitespaceWidth < 0 {
		log.Errorf("Invalid time box size for RunTime: %#v", runtime)
		whitespaceWidth = 0
	}

	log.Printf("| Total runtime: %s%s|", runTimeStr, strings.Repeat(" ", whitespaceWidth))
	log.Printf("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	log.Print()
}
