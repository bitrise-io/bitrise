package stringutil

// MaxLastChars returns the last maxCharCount characters,
//  or in case maxCharCount is more than or equal to the string's length
//  it'll just return the whole string.
func MaxLastChars(inStr string, maxCharCount int) string {
	strLen := len(inStr)
	if maxCharCount >= strLen {
		return inStr
	}
	return inStr[strLen-maxCharCount:]
}
