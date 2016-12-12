package stringutil

import (
	"bufio"
	"math"
	"strings"
)

// IndentTextWithMaxLength ...
func IndentTextWithMaxLength(text string, indent string, maxTextLineCharWidth int, isIndentFirstLine bool) string {
	if maxTextLineCharWidth < 1 {
		return ""
	}

	formattedText := ""

	addLine := func(line string) {
		isFirstLine := (formattedText == "")
		if !isFirstLine {
			formattedText = formattedText + "\n"
		}
		if isFirstLine && !isIndentFirstLine {
			formattedText = line
		} else {
			formattedText = formattedText + indent + line
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		lineLength := len(line)
		if lineLength > maxTextLineCharWidth {
			lineCnt := math.Ceil(float64(lineLength) / float64(maxTextLineCharWidth))
			for i := 0; i < int(lineCnt); i++ {
				startIdx := i * maxTextLineCharWidth
				endIdx := startIdx + maxTextLineCharWidth
				if endIdx > lineLength {
					endIdx = lineLength
				}
				addLine(line[startIdx:endIdx])
			}
		} else {
			addLine(line)
		}
	}

	return formattedText
}
