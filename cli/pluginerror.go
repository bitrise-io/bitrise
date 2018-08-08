package cli

// NewPluginError returns a pluginError with the provided text
func NewPluginError(text string) error {
	return PluginError{
		s: text,
	}
}

// PluginError ...
type PluginError struct {
	s string
}

func (e PluginError) Error() string {
	return e.s
}
