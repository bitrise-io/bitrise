package pointers

import "time"

// NewBoolPtr ...
func NewBoolPtr(val bool) *bool {
	ptrValue := new(bool)
	*ptrValue = val
	return ptrValue
}

// NewStringPtr ...
func NewStringPtr(val string) *string {
	ptrValue := new(string)
	*ptrValue = val
	return ptrValue
}

// NewTimePtr ...
func NewTimePtr(val time.Time) *time.Time {
	ptrValue := new(time.Time)
	*ptrValue = val
	return ptrValue
}

// NewIntPtr ...
func NewIntPtr(val int) *int {
	ptrValue := new(int)
	*ptrValue = val
	return ptrValue
}

// ------------------------------------------------------
// --- Safe Getters

// Bool ...
func Bool(val *bool) bool {
	return BoolWithDefault(val, false)
}

// BoolWithDefault ...
func BoolWithDefault(val *bool, defaultValue bool) bool {
	if val == nil {
		return defaultValue
	}
	return *val
}

// String ...
func String(val *string) string {
	return StringWithDefault(val, "")
}

// StringWithDefault ...
func StringWithDefault(val *string, defaultValue string) string {
	if val == nil {
		return defaultValue
	}
	return *val
}

// TimeWithDefault ...
func TimeWithDefault(val *time.Time, defaultValue time.Time) time.Time {
	if val == nil {
		return defaultValue
	}
	return *val
}

// Int ...
func Int(val *int) int {
	return IntWithDefault(val, 0)
}

// IntWithDefault ...
func IntWithDefault(val *int, defaultValue int) int {
	if val == nil {
		return defaultValue
	}
	return *val
}
