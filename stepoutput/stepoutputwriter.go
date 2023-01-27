package stepoutput

import (
	"io"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/bitrise/tools/errorfinder"
	"github.com/bitrise-io/bitrise/tools/filterwriter"
)

type Writer interface {
	Write(p []byte) (n int, err error)
	Flush() (int, error)
	ErrorMessages() []errorfinder.ErrorMessage
}

type defaultWriter struct {
	secretWriter *filterwriter.Writer
	errorWriter  *errorfinder.ErrorFinder

	writer io.Writer
}

func NewWriter(secrets []string, opts log.LoggerOpts) Writer {
	var outWriter io.Writer

	logWriter := logwriter.NewLogWriter(log.NewLogger(opts))
	outWriter = logWriter

	var secretWriter *filterwriter.Writer
	if len(secrets) > 0 {
		secretWriter = filterwriter.New(secrets, outWriter)
		outWriter = secretWriter
	}

	errorWriter := errorfinder.NewErrorFinder(outWriter, errorfinder.NewDefaultTimestampProvider())
	outWriter = errorWriter

	return defaultWriter{
		secretWriter: secretWriter,
		errorWriter:  errorWriter,
		writer:       outWriter,
	}
}

func (w defaultWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w defaultWriter) Flush() (int, error) {
	return w.secretWriter.Flush()
}

func (w defaultWriter) ErrorMessages() []errorfinder.ErrorMessage {
	return w.errorWriter.ErrorMessages()
}
