package bitrise

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/toolkits"
	"github.com/bitrise-io/go-utils/colorstring"
	log "github.com/bitrise-io/go-utils/log"
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

	titleBox := ""
	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess, models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		titleBox = fmt.Sprintf("%s", title)
		if len(titleBox) > titleBoxWidth {
			dif := len(titleBox) - titleBoxWidth
			title = stringutil.MaxFirstCharsWithDots(title, len(title)-dif)
			titleBox = fmt.Sprintf("%s", title)
		}
		break
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodeFailedSkippable:
		titleBox = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		if len(titleBox) > titleBoxWidth {
			dif := len(titleBox) - titleBoxWidth
			title = stringutil.MaxFirstCharsWithDots(title, len(title)-dif)
			titleBox = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		}
		break
	default:
		log.Errorf("Unknown result code")
		return ""
	}

	return titleBox
}

func getRunningStepHeaderMainSection(stepInfo stepmanModels.StepInfoModel, idx int) string {
	title := ""
	if stepInfo.Step.Title != nil && *stepInfo.Step.Title != "" {
		title = *stepInfo.Step.Title
	}

	content := fmt.Sprintf("| (%d) %s |", idx, title)
	charDiff := len(content) - stepRunSummaryBoxWidthInChars

	if charDiff < 0 {
		// shorter than desired - fill with space
		content = fmt.Sprintf("| (%d) %s%s |", idx, title, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedTitleWidth := len(title) - charDiff
		if trimmedTitleWidth < 4 {
			log.Errorf("Step title too long, can't present title at all! : %s", title)
		} else {
			content = fmt.Sprintf("| (%d) %s |", idx, stringutil.MaxFirstCharsWithDots(title, trimmedTitleWidth))
		}
	}
	return content
}

func getRunningStepHeaderSubSection(step stepmanModels.StepModel, stepInfo stepmanModels.StepInfoModel) string {

	idRow := ""
	{
		id := stepInfo.ID
		idRow = fmt.Sprintf("| id: %s |", id)
		charDiff := len(idRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			idRow = fmt.Sprintf("| id: %s%s |", id, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(id) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Step id too long, can't present id at all! : %s", id)
			} else {
				idRow = fmt.Sprintf("| id: %s |", stringutil.MaxFirstCharsWithDots(id, trimmedWidth))
			}
		}
	}

	versionRow := ""
	{
		version := stepInfo.Version
		versionRow = fmt.Sprintf("| version: %s |", version)
		charDiff := len(versionRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			versionRow = fmt.Sprintf("| version: %s%s |", version, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(version) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Step version too long, can't present version at all! : %s", version)
			} else {
				versionRow = fmt.Sprintf("| id: %s |", stringutil.MaxFirstCharsWithDots(version, trimmedWidth))
			}
		}
	}

	collectionRow := ""
	{
		collection := stepInfo.Library
		collectionRow = fmt.Sprintf("| collection: %s |", collection)
		charDiff := len(collectionRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			collectionRow = fmt.Sprintf("| collection: %s%s |", collection, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(collection) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Step collection too long, can't present collection at all! : %s", collection)
			} else {
				collectionRow = fmt.Sprintf("| collection: %s |", stringutil.MaxLastCharsWithDots(collection, trimmedWidth))
			}
		}
	}

	toolkitRow := ""
	{
		toolkitForStep := toolkits.ToolkitForStep(step)
		toolkitName := toolkitForStep.ToolkitName()
		toolkitRow = fmt.Sprintf("| toolkit: %s |", toolkitName)
		charDiff := len(toolkitRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			toolkitRow = fmt.Sprintf("| toolkit: %s%s |", toolkitName, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(toolkitName) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Step toolkitName too long, can't present toolkitName at all! : %s", toolkitName)
			} else {
				toolkitRow = fmt.Sprintf("| toolkit: %s |", stringutil.MaxLastCharsWithDots(toolkitName, trimmedWidth))
			}
		}
	}

	timeRow := ""
	{
		logTime := time.Now().Format(time.RFC3339)
		timeRow = fmt.Sprintf("| time: %s |", logTime)
		charDiff := len(timeRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			timeRow = fmt.Sprintf("| time: %s%s |", logTime, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			trimmedWidth := len(logTime) - charDiff
			if trimmedWidth < 4 {
				log.Errorf("Time too long, can't present time at all! : %s", logTime)
			} else {
				timeRow = fmt.Sprintf("| time: %s |", stringutil.MaxFirstCharsWithDots(logTime, trimmedWidth))
			}
		}
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", idRow, versionRow, collectionRow, toolkitRow, timeRow)
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
		break
	case models.StepRunStatusCodeFailed:
		icon = "x"
		coloringFunc = colorstring.Red
		break
	case models.StepRunStatusCodeFailedSkippable:
		icon = "!"
		coloringFunc = colorstring.Yellow
		break
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		icon = "-"
		coloringFunc = colorstring.Blue
		break
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

	runTimeStr, err := FormattedSecondsToMax8Chars(stepRunResult.RunTime)
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
		isLastLine := (i == len(words)-1)

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

// evaluatedVersion introduced for testing purposes, until StepInfoModel is updated to have the same property
func buildUpdateRow(stepInfo stepmanModels.StepInfoModel, width int, evaluatedVersion string) string {
	vstr := fmt.Sprintf("%s -> %s", stepInfo.Version, stepInfo.LatestVersion)
	if stepInfo.Version != evaluatedVersion {
		vstr = fmt.Sprintf("%s (%s) -> %s", stepInfo.Version, evaluatedVersion, stepInfo.LatestVersion)
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

	isUpdateAvailable := isUpdateAvailable(stepRunResult.StepInfo)
	updateRow := ""
	if isUpdateAvailable {
		updateRow = buildUpdateRow(stepInfo, stepRunSummaryBoxWidthInChars, stepInfo.EvaluatedVersion)
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
			releasesURL := *stepInfo.Step.SourceCodeURL + "/releases"
			content += "\n" + getRow("Release notes are available on GitHub")
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

// PrintRunningStepHeader ...
func PrintRunningStepHeader(stepInfo stepmanModels.StepInfoModel, step stepmanModels.StepModel, idx int) {
	sep := fmt.Sprintf("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	fmt.Println(sep)
	fmt.Println(getRunningStepHeaderMainSection(stepInfo, idx))
	fmt.Println(sep)
	fmt.Println(getRunningStepHeaderSubSection(step, stepInfo))
	fmt.Println(sep)
	fmt.Println("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")
}

// PrintRunningStepFooter ..
func PrintRunningStepFooter(stepRunResult models.StepRunResultsModel, isLastStepInWorkflow bool) {
	iconBoxWidth := len("   ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth
	sep := fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	fmt.Println("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")

	fmt.Println(sep)
	fmt.Println(getRunningStepFooterMainSection(stepRunResult))
	fmt.Println(sep)
	if stepRunResult.ErrorStr != "" || stepRunResult.StepInfo.GroupInfo.RemovalDate != "" || isUpdateAvailable(stepRunResult.StepInfo) {
		footerSubSection := getRunningStepFooterSubSection(stepRunResult)
		if footerSubSection != "" {
			fmt.Println(footerSubSection)
			fmt.Println(sep)
		}
	}

	if !isLastStepInWorkflow {
		fmt.Println()
		fmt.Println(strings.Repeat(" ", 42) + "▼")
		fmt.Println()
	}
}

// PrintRunningWorkflow ...
func PrintRunningWorkflow(title string) {
	fmt.Println()
	log.Printf("%s %s", colorstring.Blue("Switching to workflow:"), title)
	fmt.Println()
}

// PrintSummary ...
func PrintSummary(buildRunResults models.BuildRunResultsModel) {
	iconBoxWidth := len("   ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth

	fmt.Println()
	fmt.Println()
	fmt.Printf("+%s+\n", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))
	whitespaceWidth := (stepRunSummaryBoxWidthInChars - 2 - len("bitrise summary ")) / 2
	fmt.Printf("|%sbitrise summary %s|\n", strings.Repeat(" ", whitespaceWidth), strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	whitespaceWidth = stepRunSummaryBoxWidthInChars - len("|   | title") - len("| time (s) |")
	fmt.Printf("|   | title%s| time (s) |\n", strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	orderedResults := buildRunResults.OrderedResults()
	tmpTime := time.Time{}
	for _, stepRunResult := range orderedResults {
		tmpTime = tmpTime.Add(stepRunResult.RunTime)
		fmt.Println(getRunningStepFooterMainSection(stepRunResult))
		fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))
		if stepRunResult.ErrorStr != "" || stepRunResult.StepInfo.GroupInfo.RemovalDate != "" || isUpdateAvailable(stepRunResult.StepInfo) {
			footerSubSection := getRunningStepFooterSubSection(stepRunResult)
			if footerSubSection != "" {
				fmt.Println(footerSubSection)
				fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))
			}
		}
	}
	runtime := tmpTime.Sub(time.Time{})

	runTimeStr, err := FormattedSecondsToMax8Chars(runtime)
	if err != nil {
		log.Errorf("Failed to format time, error: %s", err)
		runTimeStr = "999+ hour"
	}

	whitespaceWidth = stepRunSummaryBoxWidthInChars - len(fmt.Sprintf("| Total runtime: %s|", runTimeStr))
	if whitespaceWidth < 0 {
		log.Errorf("Invalid time box size for RunTime: %#v", runtime)
		whitespaceWidth = 0
	}

	fmt.Printf("| Total runtime: %s%s|\n", runTimeStr, strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+\n", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	fmt.Println()
}
