package logwriter

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/bitrise-io/go-utils/v2/advancedlog/corelog"
)

var ansiEscapeCodeToLevel = map[corelog.ANSIColorCode]corelog.Level{
	corelog.RedCode:     corelog.ErrorLevel,
	corelog.YellowCode:  corelog.WarnLevel,
	corelog.BlueCode:    corelog.InfoLevel,
	corelog.GreenCode:   corelog.DoneLevel,
	corelog.MagentaCode: corelog.DebugLevel,
}

// convertColoredString determines the log level of the given message and removes related ANSI escape codes from it.
// Messages without a log level are returned untouched.
// A message is considered a message with a log level if:
// - starts with a color code
// - the first trailing non-whitespace characters have to be part of the reset-color code
// - contains exactly one color and reset code pair.
func convertColoredString(message string) (corelog.Level, string) {
	// We need to remove all the possible noise from the end as we need remove the reset ansi code from the end
	trimmedMessage := strings.TrimRightFunc(message, unicode.IsSpace)

	// If the message has more than one color then let the website do the coloring and do not modify the message
	if hasMoreThanOneColor(trimmedMessage) {
		return corelog.NormalLevel, message
	}

	// Some messages have the starting color but do not have the reset code at the end. Ignore these.
	if !strings.HasSuffix(trimmedMessage, string(corelog.ResetCode)) {
		return corelog.NormalLevel, message
	}

	for code, level := range ansiEscapeCodeToLevel {
		if strings.HasPrefix(message, string(code)) {
			message = strings.TrimPrefix(message, string(code))
			message = strings.Replace(message, string(corelog.ResetCode), "", 1)
			return level, message
		}
	}

	return corelog.NormalLevel, message
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
