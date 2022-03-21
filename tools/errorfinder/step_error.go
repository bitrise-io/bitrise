package errorfinder

// StepError ...
type StepError struct {
	Message string
	Err     error
}

// Unwrap ...
func (e *StepError) Unwrap() error { return e.Err }

func (e *StepError) Error() string { return e.Message }
