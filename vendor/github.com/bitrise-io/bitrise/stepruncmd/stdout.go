package stepruncmd

import (
	"io"

	"github.com/bitrise-io/bitrise/stepruncmd/errorfinder"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/redactwriter"
)

type StdoutWriter struct {
	writer io.Writer

	redactWriter *redactwriter.Writer
	errorWriter  *errorfinder.ErrorFinder
	destWriter   io.Writer
}

func NewStdoutWriter(secrets []string, dest io.Writer, logger log.Logger) StdoutWriter {
	var outWriter io.Writer
	outWriter = dest

	errorWriter := errorfinder.NewErrorFinder(outWriter)
	outWriter = errorWriter

	var redactWriter *redactwriter.Writer
	if len(secrets) > 0 {
		redactWriter = redactwriter.New(secrets, outWriter, logger)
		outWriter = redactWriter
	}

	return StdoutWriter{
		writer: outWriter,

		redactWriter: redactWriter,
		errorWriter:  errorWriter,
		destWriter:   dest,
	}
}

func (w StdoutWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w StdoutWriter) Close() error {
	if w.redactWriter != nil {
		if err := w.redactWriter.Close(); err != nil {
			return err
		}
	}

	if err := w.errorWriter.Close(); err != nil {
		return err
	}

	if writeCloser, ok := w.destWriter.(io.WriteCloser); ok {
		if err := writeCloser.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (w StdoutWriter) ErrorMessages() []string {
	return w.errorWriter.ErrorMessages()
}
