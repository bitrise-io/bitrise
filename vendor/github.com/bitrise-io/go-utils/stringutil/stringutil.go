package stringutil

import "fmt"

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
