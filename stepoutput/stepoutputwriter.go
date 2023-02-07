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
	writer io.Writer

	secretWriter   *filterwriter.Writer
	errorWriter    *errorfinder.ErrorFinder
	logLevelWriter *logwriter.LogLevelWriter
}

func NewWriter(secrets []string, opts log.LoggerOpts) Writer {
	var outWriter io.Writer

	logLevelWriter := logwriter.NewLogLevelWriter(log.NewLogger(opts))
	outWriter = logLevelWriter

	errorWriter := errorfinder.NewErrorFinder(outWriter, opts.TimeProvider)
	outWriter = errorWriter

	var secretWriter *filterwriter.Writer
	if len(secrets) > 0 {
		secretWriter = filterwriter.New(secrets, outWriter)
		outWriter = secretWriter
	}

	return defaultWriter{
		writer: outWriter,

		secretWriter:   secretWriter,
		errorWriter:    errorWriter,
		logLevelWriter: logLevelWriter,
	}
}

func (w defaultWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w defaultWriter) Flush() (int, error) {
	if w.secretWriter != nil {
		n, err := w.secretWriter.Flush()
		if err != nil {
			return n, err
		}

	}

	if w.logLevelWriter != nil {
		return w.logLevelWriter.Flush()
	}
	
	return 0, nil
}

func (w defaultWriter) ErrorMessages() []errorfinder.ErrorMessage {
	if w.errorWriter != nil {
		return w.errorWriter.ErrorMessages()
	}
	return nil
}
