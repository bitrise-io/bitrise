package cli

import "fmt"

type EnvVarValueTooLargeError struct {
	Key           string
	ValueSizeInKB float64
	MaxSizeInKB   float64
}

func NewEnvVarValueTooLargeError(key string, valueSizeInKB, maxSizeInKB float64) error {
	return EnvVarValueTooLargeError{
		Key:           key,
		ValueSizeInKB: valueSizeInKB,
		MaxSizeInKB:   maxSizeInKB,
	}
}

func (err EnvVarValueTooLargeError) Error() string {
	return fmt.Sprintf("env var (%s) value is too large (%#v KB), max allowed size: %#v KB", err.Key, err.ValueSizeInKB, err.MaxSizeInKB)
}

type EnvVarListTooLargeError struct {
	EnvListSizeInKB float64
	MaxSizeInKB     float64
}

func NewEnvVarListTooLargeError(envListSizeInKB, maxSizeInKB float64) EnvVarListTooLargeError {
	return EnvVarListTooLargeError{
		EnvListSizeInKB: envListSizeInKB,
		MaxSizeInKB:     maxSizeInKB,
	}
}

func (e EnvVarListTooLargeError) Error() string {
	return fmt.Sprintf("env var list is too large (%#v KB), max allowed size: %#v KB", e.EnvListSizeInKB, e.MaxSizeInKB)
}
