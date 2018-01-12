package stringutil

import (
	"bufio"
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
	sLen := len(inStr)
	bStr := []rune(inStr)
	trimIndex := 0

	if maxCharCount >= sLen {
		return inStr
	}

	if trimmAtStart {
		bStr = append([]rune(nil), bStr[(sLen-maxCharCount):]...)
	} else {
		bStr = append([]rune(nil), bStr[:maxCharCount]...)
		trimIndex = maxCharCount - 1
	}
	if appendDots {
		bStr[trimIndex] = '…'
	}
	return string(bStr)
}
