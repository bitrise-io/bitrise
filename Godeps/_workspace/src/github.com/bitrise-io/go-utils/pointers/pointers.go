package pointers

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
