package stringutil

import (
	"bufio"
	"fmt"
	"strings"
)

// ReadFirstLine ...
func ReadFirstLine(s string, isIgnoreLeadingEmptyLines bool) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	firstLine := ""
	for scanner.Scan() {
		firstLine = scanner.Text()
		if !isIgnoreLeadingEmptyLines || firstLine != "" {
			break
		}
	}
	return firstLine
}

// CaseInsensitiveEquals ...
func CaseInsensitiveEquals(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	return a == b
}

// CaseInsensitiveContains ...
func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// MaxLastChars returns the last maxCharCount characters,
//  or in case maxCharCount is more than or equal to the string's length
//  it'll just return the whole string.
func MaxLastChars(inStr string, maxCharCount int) string {
	return genericTrim(inStr, maxCharCount, true, false)
}

// MaxLastCharsWithDots ...
func MaxLastCharsWithDots(inStr string, maxCharCount int) string {
	return genericTrim(inStr, maxCharCount, true, true)
}

// MaxFirstChars ...
func MaxFirstChars(inStr string, maxCharCount int) string {
	return genericTrim(inStr, maxCharCount, false, false)
}

// MaxFirstCharsWithDots ...
func MaxFirstCharsWithDots(inStr string, maxCharCount int) string {
	return genericTrim(inStr, maxCharCount, false, true)
}

func genericTrim(inStr string, maxCharCount int, trimmAtStart, appendDots bool) string {
	retStr := inStr
	strLen := len(inStr)

	if maxCharCount >= strLen {
		return retStr
	}

	if appendDots {
		if maxCharCount < 4 {
			fmt.Println("Append dots mode, but string length < 4")
			return ""
		}
	}
	if trimmAtStart {
		if appendDots {
			retStr = inStr[strLen-(maxCharCount-3):]
		} else {
			retStr = inStr[strLen-maxCharCount:]
		}
	} else {
		if appendDots {
			retStr = inStr[:maxCharCount-3]
		} else {
			retStr = inStr[:maxCharCount]
		}
	}

	if appendDots {
		if trimmAtStart {
			retStr = "..." + retStr
		} else {
			retStr = retStr + "..."
		}
	}
	return retStr
}

// LastNLines ...
func LastNLines(s string, n int) string {
	trimmed := strings.Trim(s, "\n")
	splitted := strings.Split(trimmed, "\n")

	if len(splitted) >= n {
		splitted = splitted[len(splitted)-n:]
	}

	return strings.Join(splitted, "\n")
}
