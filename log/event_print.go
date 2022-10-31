package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/log/corelog"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/utils"
	"github.com/bitrise-io/go-utils/stringutil"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 80
	footerIconBoxWidth            = 3
	footerExecutionTimeBoxWidth   = 10
	// the -4 value is the four `|` separator characters in the line
	footerTitleBoxWidth   = stepRunSummaryBoxWidthInChars - footerIconBoxWidth - footerExecutionTimeBoxWidth - 4
	deprecatedPrefix      = "[Deprecated]"
	missingUrlPlaceholder = "Not provided"
	removalDateTitle      = "Removal date:"
	removalNotesTitle     = "Removal notes:"
	updateAvailableTitle  = "Update available"
)

// This is the main entry point to generate the started event console log lines.
func generateStepStartedHeaderLines(params StepStartedParams) []string {
	separatorContentWidth := stepRunSummaryBoxWidthInChars - 2
	separator := fmt.Sprintf("+%s+", strings.Repeat("-", separatorContentWidth))
	stepLogStartIndicator := fmt.Sprintf("|%s|", strings.Repeat(" ", separatorContentWidth))

	var lines []string
	lines = append(lines, separator)
	lines = append(lines, getHeaderTitle(params.Position, params.Title))
	lines = append(lines, separator)
	lines = append(lines, getHeaderSubsection("id", params.Id))
	lines = append(lines, getHeaderSubsection("version", params.Version))
	lines = append(lines, getHeaderSubsection("collection", params.Collection))
	lines = append(lines, getHeaderSubsection("toolkit", params.Toolkit))
	lines = append(lines, getHeaderSubsection("time", params.StartTime))
	lines = append(lines, separator)
	lines = append(lines, stepLogStartIndicator)

	return lines
}

func getHeaderTitle(position int, title string) string {
	content := fmt.Sprintf("(%d) %s", position, title)
	return getHeaderLine(content)
}

func getHeaderSubsection(key, value string) string {
	content := fmt.Sprintf("%s: %s", key, value)
	return getHeaderLine(content)
}

func getHeaderLine(content string) string {
	// The available space is 4 char less because if the `| ` prefix and ` |` suffix every line has.
	contentMaxWidth := stepRunSummaryBoxWidthInChars - 4
	return widthConstrainedStringWithBorder(content, contentMaxWidth)
}

func widthConstrainedStringWithBorder(content string, width int) string {
	return fmt.Sprintf("| %s |", widthConstrainedString(content, width))
}

func widthConstrainedString(content string, width int) string {
	widthDiff := len(content) - width

	if widthDiff < 0 {
		return fmt.Sprintf("%s%s", content, strings.Repeat(" ", -widthDiff))
	}

	if widthDiff == 0 {
		return content
	}

	return stringutil.MaxFirstCharsWithDots(content, width)
}

// This is the main entry point to generate the finished event console log lines.
func generateStepFinishedFooterLines(params StepFinishedParams) []string {
	mainSeparator := fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", footerIconBoxWidth), strings.Repeat("-", footerTitleBoxWidth), strings.Repeat("-", footerExecutionTimeBoxWidth))
	sectionSeparator := fmt.Sprintf("|%s|", strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2))

	deprecated := false
	if params.Deprecation != nil {
		deprecated = params.Deprecation.RemovalDate != "" || params.Deprecation.Note != ""
	}

	var lines []string
	lines = append(lines, sectionSeparator)
	lines = append(lines, mainSeparator)
	lines = append(lines, getSummaryFooterRow(params.InternalStatus, params.Title, params.StatusReason, params.RunTime, deprecated))
	lines = append(lines, mainSeparator)

	hasPreviousSection := false

	if len(params.Errors) > 0 {
		hasPreviousSection = true
		lines = append(lines, getIssueAndSourceSection(params.SupportURL, params.SourceCodeURL)...)
	}

	if params.Update != nil {
		if hasPreviousSection {
			lines = append(lines, sectionSeparator)
		}
		hasPreviousSection = true
		lines = append(lines, getUpdateSection(*params.Update)...)
	}

	if deprecated {
		if hasPreviousSection {
			lines = append(lines, sectionSeparator)
		}
		hasPreviousSection = true
		lines = append(lines, getDeprecationSection(*params.Deprecation)...)
	}

	if hasPreviousSection {
		lines = append(lines, mainSeparator)
	}

	if !params.LastStep {
		lines = append(lines, "")
		lines = append(lines, strings.Repeat(" ", stepRunSummaryBoxWidthInChars/2)+"▼")
		lines = append(lines, "")
	}

	return lines
}

func getSummaryFooterRow(status int, title, reason string, duration int64, deprecated bool) string {
	icon, level := transformStatusToIconAndLevel(status)
	footerTitle := getFooterTitle(level, title, reason, deprecated, footerTitleBoxWidth)
	executionTime := getFooterExecutionTime(duration)

	return fmt.Sprintf("|%s|%s|%s|", icon, footerTitle, executionTime)
}

func transformStatusToIconAndLevel(status int) (string, corelog.Level) {
	var icon string
	var level corelog.Level

	switch status {
	case models.StepRunStatusCodeSuccess:
		icon = "✓"
		level = corelog.DoneLevel
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodePreparationFailed:
		icon = "x"
		level = corelog.ErrorLevel
	case models.StepRunStatusAbortedWithCustomTimeout, models.StepRunStatusAbortedWithNoOutputTimeout:
		icon = "/"
		level = corelog.ErrorLevel
	case models.StepRunStatusCodeFailedSkippable:
		icon = "!"
		level = corelog.WarnLevel
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		icon = "-"
		level = corelog.InfoLevel
	default:
		icon = " "
		level = corelog.NormalLevel
	}

	return fmt.Sprintf(" %s ", corelog.AddColor(level, icon)), level
}

// The title will be modified in two different ways:
// - Adding a deprecation prefix to the title
// - Appending a status reason to the end
// When both the deprivation prefix and status text is added to the title then in some cases it can be really long. We
// have a width constraint and in such cases the title would be truncated at the end which means we would lose the status
// reason information.
// This implementation leaves the prefix and suffix intact so they are always visible and instead it truncates just the
// step title.
func getFooterTitle(level corelog.Level, title, reason string, deprecated bool, width int) string {
	// Deduct the leading and trailing space from the available width
	actualWidth := width - 2
	availableTitleWidth := actualWidth
	if deprecated {
		// The +1 is for the space between the prefix and title
		availableTitleWidth -= len(deprecatedPrefix) + 1
	}
	if reason != "" {
		// The +1 is the space and the +2 the parenthesis
		availableTitleWidth -= len(reason) + 1 + 2
	}
	if len(title) > availableTitleWidth {
		title = stringutil.MaxFirstCharsWithDots(title, availableTitleWidth)
	}

	if reason != "" {
		title = fmt.Sprintf("%s (%s)", title, reason)
	}

	if deprecated {
		title = strings.TrimPrefix(title, deprecatedPrefix)
		title = strings.TrimSpace(title)

		// Deduct the deprecated prefix length and the space between them to only shorten the title
		actualWidth = actualWidth - len(deprecatedPrefix) - 1
		widthConstrainedTitle := widthConstrainedString(title, actualWidth)
		title = fmt.Sprintf("%s %s", corelog.AddColor(corelog.ErrorLevel, deprecatedPrefix), corelog.AddColor(level, widthConstrainedTitle))
	} else {
		title = widthConstrainedString(title, actualWidth)
		title = corelog.AddColor(level, title)
	}

	return fmt.Sprintf(" %s ", title)
}

func getFooterExecutionTime(duration int64) string {
	// Duration is in milliseconds but the formatter expects a time.Duration type which represents the time in nanoseconds.
	time := time.Duration(duration * int64(time.Millisecond))
	executionTime, err := utils.FormattedSecondsToMax8Chars(time)
	if err != nil {
		// The above FormattedSecondsToMax8Chars function only throws a single error in case the duration is more than
		// 999 hours. In this case we use a 9 char long duration string so we only need to add a space at the beginning.
		return " 999+ hour"
	}
	return fmt.Sprintf(" %s ", executionTime)
}

func getIssueAndSourceSection(issueUrl, sourceUrl string) []string {
	if issueUrl == "" {
		issueUrl = missingUrlPlaceholder
	}
	if sourceUrl == "" {
		sourceUrl = missingUrlPlaceholder
	}

	contentMaxWidth := stepRunSummaryBoxWidthInChars - 4
	issueRow := widthConstrainedStringWithBorder(fmt.Sprintf("Issue tracker: %s", issueUrl), contentMaxWidth)
	sourceRow := widthConstrainedStringWithBorder(fmt.Sprintf("Source: %s", sourceUrl), contentMaxWidth)

	coloredPlaceholder := corelog.AddColor(corelog.WarnLevel, missingUrlPlaceholder)
	issueRow = strings.ReplaceAll(issueRow, missingUrlPlaceholder, coloredPlaceholder)
	sourceRow = strings.ReplaceAll(sourceRow, missingUrlPlaceholder, coloredPlaceholder)

	return []string{issueRow, sourceRow}
}

func getUpdateSection(params StepUpdate) []string {
	versionInfo := fmt.Sprintf("%s -> %s", params.ResolvedVersion, params.LatestVersion)
	if params.OriginalVersion != params.ResolvedVersion {
		versionInfo = fmt.Sprintf("%s (%s) -> %s", params.OriginalVersion, params.ResolvedVersion, params.LatestVersion)
	}

	contentMaxWidth := stepRunSummaryBoxWidthInChars - 4
	updateRow := fmt.Sprintf("%s: %s", updateAvailableTitle, versionInfo)

	if len(updateRow) > contentMaxWidth {
		updateRow = fmt.Sprintf("%s: -> %s", updateAvailableTitle, params.LatestVersion)

		if len(updateRow) > contentMaxWidth {
			updateRow = fmt.Sprintf("%s!", updateAvailableTitle)
		}
	}

	rows := []string{widthConstrainedStringWithBorder(updateRow, contentMaxWidth)}

	if params.ReleasesURL != "" {
		rows = append(rows, widthConstrainedStringWithBorder("Release notes are available below", contentMaxWidth))
		rows = append(rows, widthConstrainedStringWithBorder(params.ReleasesURL, contentMaxWidth))
	}

	return rows
}

func getDeprecationSection(params StepDeprecation) []string {
	var rows []string
	contentMaxWidth := stepRunSummaryBoxWidthInChars - 4
	removalDateRow := widthConstrainedStringWithBorder(fmt.Sprintf("%s %s", removalDateTitle, params.RemovalDate), contentMaxWidth)
	removalDateRow = strings.ReplaceAll(removalDateRow, removalDateTitle, corelog.AddColor(corelog.ErrorLevel, removalDateTitle))
	rows = append(rows, removalDateRow)

	notesLines := reformatIntoLines(fmt.Sprintf("%s %s", removalNotesTitle, params.Note), contentMaxWidth)
	for i, line := range notesLines {
		line = widthConstrainedStringWithBorder(line, contentMaxWidth)

		if i == 0 {
			line = strings.ReplaceAll(line, removalNotesTitle, corelog.AddColor(corelog.ErrorLevel, removalNotesTitle))
		}

		rows = append(rows, line)
	}

	return rows
}

func reformatIntoLines(text string, lineWidth int) []string {
	text = strings.ReplaceAll(text, "\n", "")
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	firstWord := words[0]
	currentLine := firstWord
	remainingSpace := lineWidth - len(firstWord)

	for _, word := range words[1:] {
		wordLengthWithSpace := len(word) + 1

		if wordLengthWithSpace < remainingSpace {
			currentLine += " " + word
			remainingSpace -= wordLengthWithSpace
		} else {
			lines = append(lines, currentLine)
			currentLine = word
			remainingSpace = lineWidth - len(currentLine)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
