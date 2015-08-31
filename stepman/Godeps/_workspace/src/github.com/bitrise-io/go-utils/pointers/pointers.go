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
