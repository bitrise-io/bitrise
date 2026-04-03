package parseutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseBool parses a string representation of a boolean value.
//
// It accepts the following inputs (case-insensitive, whitespace trimmed):
//   - Custom values: "yes", "y" (true), "no", "n" (false)
//   - Standard values: "true", "t", "1" (true), "false", "f", "0" (false)
//
// Returns an error if the input is empty or cannot be parsed as a boolean.
func ParseBool(input string) (bool, error) {
	// Validation
	if input == "" {
		return false, errors.New("no string to parse")
	}

	// Normalization
	normalized := strings.ToLower(strings.TrimSpace(input))

	// Custom parsing
	switch normalized {
	case "yes", "y":
		return true, nil
	case "no", "n":
		return false, nil
	}

	// Delegate to stdlib
	return strconv.ParseBool(normalized)
}

// StringFrom converts any value to its string representation.
//
// If the value is already a string, it returns it directly for efficiency.
// Otherwise, it uses fmt.Sprintf with the %v verb to convert the value.
//
// This function always returns a string and never returns an error.
//
// Example:
//
//	StringFrom("hello")    // "hello"
//	StringFrom(42)         // "42"
//	StringFrom(true)       // "true"
//	StringFrom(3.14)       // "3.14"
func StringFrom(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

// StringPtrFrom converts any value to its string representation and returns a pointer.
//
// This is a convenience function that calls StringFrom and returns a pointer to the result.
//
// Example:
//
//	StringPtrFrom(42)      // pointer to "42"
//	StringPtrFrom("test")  // pointer to "test"
func StringPtrFrom(value interface{}) *string {
	result := StringFrom(value)
	return &result
}

// BoolFrom attempts to convert a value to a boolean.
//
// It returns (result, true) on success, or (false, false) if it fails.
//
// Conversion rules:
//   - If the value is already a bool, it returns it directly
//   - If the value can be converted to a string and parsed by ParseBool, the parsed result is returned
//   - Otherwise, it returns (false, false)
func BoolFrom(value interface{}) (bool, bool) {
	// Fast path: if already a bool, return it directly
	if b, ok := value.(bool); ok {
		return b, true
	}

	// Try to convert to string and parse
	str := StringFrom(value)
	result, err := ParseBool(str)
	if err != nil {
		return false, false
	}
	return result, true
}

// BoolPtrFrom attempts to convert a value to a boolean pointer.
//
// It returns (pointer, true) on success, or (nil, false) if it fails.
// This is a convenience function wrapping BoolFrom.
//
// Example:
//
//	BoolPtrFrom(true)      // (pointer to true, true)
//	BoolPtrFrom("yes")     // (pointer to true, true)
//	BoolPtrFrom("invalid") // (nil, false)
func BoolPtrFrom(value interface{}) (*bool, bool) {
	result, ok := BoolFrom(value)
	if !ok {
		return nil, false
	}
	return &result, true
}
