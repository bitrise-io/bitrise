package parseutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/pointers"
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
func CastToString(value interface{}) string {
	casted, ok := value.(string)

	if !ok {
		castedStr := fmt.Sprintf("%v", value)
		casted = castedStr
	}

	return casted
}

// CastToStringPtr ...
func CastToStringPtr(value interface{}) *string {
	castedValue := CastToString(value)
	return pointers.NewStringPtr(castedValue)
}

// CastToBool ...
func CastToBool(value interface{}) (bool, bool) {
	casted, ok := value.(bool)

	if !ok {
		castedStr := CastToString(value)

		castedBool, err := ParseBool(castedStr)
		if err != nil {
			return false, false
		}

		casted = castedBool
	}

	return casted, true
}

// CastToBoolPtr ...
func CastToBoolPtr(value interface{}) (*bool, bool) {
	castedValue, ok := CastToBool(value)
	if !ok {
		return nil, false
	}
	return pointers.NewBoolPtr(castedValue), true
}

// CastToMapStringInterface ...
func CastToMapStringInterface(value interface{}) (map[string]interface{}, bool) {
	castedValue, ok := value.(map[interface{}]interface{})
	desiredMap := map[string]interface{}{}
	for key, value := range castedValue {
		keyStr, ok := key.(string)
		if !ok {
			return map[string]interface{}{}, false
		}
		desiredMap[keyStr] = value
	}
	return desiredMap, ok
}

// CastToMapStringInterfacePtr ...
func CastToMapStringInterfacePtr(value interface{}) (*map[string]interface{}, bool) {
	casted, ok := CastToMapStringInterface(value)
	if !ok {
		return nil, false
	}
	return pointers.NewMapStringInterfacePtr(casted), true
}
