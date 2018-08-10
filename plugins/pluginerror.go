package plugins

import "fmt"

// NewNotInstalledError returns a not installed error with the provided plugin
func NewNotInstalledError(plugin string) error {
	return NotInstalledError{
		plugin: plugin,
	}
}

// NotInstalledError ...
type NotInstalledError struct {
	plugin string
}

func (e NotInstalledError) Error() string {
	return fmt.Sprintf("%s not installed", e.plugin)
}
