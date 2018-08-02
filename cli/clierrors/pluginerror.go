package clierrors

// NewPluginError returns a pluginError with the provided text and retry option
func NewPluginError(text string, retry bool) error {
	return PluginError{
		s:     text,
		Retry: retry,
	}
}

// PluginError ...
type PluginError struct {
	s     string
	Retry bool
}

func (e PluginError) Error() string {
	return e.s
}
