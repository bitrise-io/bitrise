package stepoutput

import (
	"io"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/bitrise/tools/errorfinder"
	"github.com/bitrise-io/bitrise/tools/filterwriter"
)

type Writer struct {
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

	return Writer{
		writer: outWriter,

		secretWriter:   secretWriter,
		errorWriter:    errorWriter,
		logLevelWriter: logLevelWriter,
	}
}

func (w Writer) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w Writer) Close() error {
	if w.secretWriter != nil {
		if err := w.secretWriter.Close(); err != nil {
			return err
		}
	}

	if err := w.errorWriter.Close(); err != nil {
		return err
	}

	return w.logLevelWriter.Close()
}

func (w Writer) ErrorMessages() []string {
	return w.errorWriter.ErrorMessages()
}
