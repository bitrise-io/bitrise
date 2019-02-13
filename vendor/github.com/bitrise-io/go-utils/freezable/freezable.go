package freezable

import (
	"fmt"
)

//
// This package implements a `freeze` function,
//  similar to Ruby's `freeze` method: http://ruby-doc.org/core-2.3.0/Object.html#method-i-freeze
// Once an object is fozen you can't unfreeze it, and `Set` will return an error
//  if called on a frozen object.
// You can check whether the object is frozen with the `IsFrozen` function.
//

// --- String ----------------------------

// String ...
type String struct {
	data     *string
	isFrozen bool
}

// Set ...
func (freezableObj *String) Set(s string) error {
	if freezableObj.isFrozen {
		return fmt.Errorf("freezable.String: Object is already frozen. (Current value: %s) (New value was: %s)",
			freezableObj.String(), s)
	}

	freezableObj.data = &s
	return nil
}

// Freeze ...
func (freezableObj *String) Freeze() {
	freezableObj.isFrozen = true
}

// IsFrozen ...
func (freezableObj *String) IsFrozen() bool {
	return freezableObj.isFrozen
}

// Get ...
func (freezableObj String) Get() string {
	if freezableObj.data == nil {
		return ""
	}
	return *freezableObj.data
}

// String ...
func (freezableObj String) String() string {
	return freezableObj.Get()
}

// --- StringSlice ----------------------------

// StringSlice ...
type StringSlice struct {
	data     *[]string
	isFrozen bool
}

// Set ...
func (freezableObj *StringSlice) Set(s []string) error {
	if freezableObj.isFrozen {
		return fmt.Errorf("freezable.StringSlice: Object is already frozen. (Current value: %s) (New value was: %s)",
			freezableObj.Get(), s)
	}

	freezableObj.data = &s
	return nil
}

// Freeze ...
func (freezableObj *StringSlice) Freeze() {
	freezableObj.isFrozen = true
}

// IsFrozen ...
func (freezableObj *StringSlice) IsFrozen() bool {
	return freezableObj.isFrozen
}

// Get ...
func (freezableObj StringSlice) Get() []string {
	if freezableObj.data == nil {
		return []string{}
	}
	return *freezableObj.data
}

// String ...
func (freezableObj StringSlice) String() string {
	return fmt.Sprintf("%s", freezableObj.Get())
}
