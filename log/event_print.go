package log

import (
	"fmt"
	"github.com/bitrise-io/go-utils/stringutil"
	"strings"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 80
)

func generateStepStartedHeaderLines(params StepStartedParams) []string {
	separatorContentWidth := stepRunSummaryBoxWidthInChars - 2
	separator := fmt.Sprintf("+%s+", strings.Repeat("-", separatorContentWidth))
	stepLogStartIndicator := fmt.Sprintf("|%s|", strings.Repeat(" ", separatorContentWidth))

	var lines []string
	lines = append(lines, separator)
	lines = append(lines, getHeaderTitle(params.Position, params.IdVersion))
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
	widthDiff := len(content) - contentMaxWidth

	if widthDiff < 0 {
		return fmt.Sprintf("| %s%s |", content, strings.Repeat(" ", -widthDiff))
	}

	if widthDiff == 0 {
		return fmt.Sprintf("| %s |", content)
	}

	return fmt.Sprintf("| %s |", stringutil.MaxFirstCharsWithDots(content, contentMaxWidth))
}
