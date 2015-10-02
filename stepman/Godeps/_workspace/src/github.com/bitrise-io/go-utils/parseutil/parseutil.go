package parseutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseBool ...
func ParseBool(userInputStr string) (bool, error) {
	if userInputStr == "" {
		return false, errors.New("No string to parse")
	}
	userInputStr = strings.TrimSpace(userInputStr)

	lowercased := strings.ToLower(userInputStr)
	if lowercased == "yes" || lowercased == "y" {
		return true, nil
	}
	if lowercased == "no" || lowercased == "n" {
		return false, nil
	}
	return strconv.ParseBool(lowercased)
}

// CastToString ...
func CastToString(v interface{}) string {
	value := fmt.Sprintf("%v", v)
	return value
}
