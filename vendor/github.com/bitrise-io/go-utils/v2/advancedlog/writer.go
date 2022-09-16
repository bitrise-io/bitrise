package logger

import (
	"io"
)

type outputInterceptor struct {
	callback func(text string)
}

func newOutputInterceptor(callback func(text string)) outputInterceptor {
	return outputInterceptor{
		callback: callback,
	}
}

func (o outputInterceptor) Write(p []byte) (n int, err error) {
	o.callback(string(p))

	return len(p), nil
}

// LogWriter ...
type LogWriter struct {
	Stdout io.Writer
	Stderr io.Writer
}

// NewLogWriter ...
func NewLogWriter(producer Producer, callback func(producer Producer, level Level, message string)) LogWriter {
	return LogWriter{
		Stdout: newOutputInterceptor(func(text string) {
			level, message := convertColoredString(text)
			callback(producer, level, message)
		}),
		Stderr: newOutputInterceptor(func(text string) {
			callback(producer, ErrorLevel, text)
		}),
	}
}
