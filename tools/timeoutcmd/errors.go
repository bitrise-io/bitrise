package timeoutcmd

import (
	"fmt"
	"time"
)

type TimeoutError struct {
	Timeout time.Duration
}

func NewTimeoutError(timeout time.Duration) *TimeoutError {
	return &TimeoutError{
		Timeout: timeout,
	}
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("timed out after %s", e.Timeout)
}

type NoOutputTimeoutError struct {
	Timeout time.Duration
}

func NewNoOutputTimeout(timeout time.Duration) *NoOutputTimeoutError {
	return &NoOutputTimeoutError{
		Timeout: timeout,
	}
}

func (e NoOutputTimeoutError) Error() string {
	return fmt.Sprintf("timed out, as no output was received for %s", e.Timeout)
}
