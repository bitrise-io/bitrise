package logger

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const resetCode = "\u001b[0m"

var ansiEscapeCodeToLevel = map[string]Level{
	"\u001b[31;1m": ErrorLevel,
	"\u001b[33;1m": WarnLevel,
	"\u001b[34;1m": InfoLevel,
	"\u001b[32;1m": DoneLevel,
	"\u001b[35;1m": DebugLevel,
}

func convertColoredString(message string) (Level, string) {
	logLevel := NormalLevel

	// We need to remove all the possible noise from the end as we need remove the reset ansi code from the end
	message = strings.TrimRightFunc(message, unicode.IsSpace)

	// If the message has more than one color then let the website do the coloring and do not modify the message
	if hasMoreThanOneColor(message) {
		return logLevel, message
	}

	// Some messages have the starting color but do not have the reset code at the end. Ignore these.
	if !strings.HasSuffix(message, resetCode) {
		return logLevel, message
	}

	for code, level := range ansiEscapeCodeToLevel {
		if strings.HasPrefix(message, code) {
			logLevel = level
			message = strings.TrimPrefix(message, code)
			message = strings.TrimSuffix(message, resetCode)

			break
		}
	}

	return logLevel, message
}

func hasMoreThanOneColor(message string) bool {
	r, err := regexp.Compile(`(\\u001b)|(\\x1b)\[.*?m`)
	if err != nil {
		return true
	}

	// The message has to be converted back to ascii characters otherwise the regex for the ansi code will not match.
	matches := r.FindAllString(strconv.QuoteToASCII(message), -1)

	var filteredMatches []string
	for _, match := range matches {
		// In this scenario the reset color does not count as a color so the additional removal. The Go regexp package
		// does not support the negative look-ahead which could ignore certain things right in the regexp.
		if !strings.Contains(match, "[0m") {
			filteredMatches = append(filteredMatches, match)
		}
	}

	return len(filteredMatches) > 1
}
