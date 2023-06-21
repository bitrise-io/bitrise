package stepruncmd

import (
	"io"

	"github.com/bitrise-io/bitrise/stepruncmd/errorfinder"
	"github.com/bitrise-io/bitrise/stepruncmd/filterwriter"
)

type StdoutWriter struct {
	writer io.Writer

	secretWriter *filterwriter.Writer
	errorWriter  *errorfinder.ErrorFinder
	destWriter   io.Writer
}

func NewStdoutWriter(secrets []string, dest io.Writer) StdoutWriter {
	var outWriter io.Writer
	outWriter = dest

	errorWriter := errorfinder.NewErrorFinder(outWriter)
	outWriter = errorWriter

	var secretWriter *filterwriter.Writer
	if len(secrets) > 0 {
		secretWriter = filterwriter.New(secrets, outWriter)
		outWriter = secretWriter
	}

	return StdoutWriter{
		writer: outWriter,

		secretWriter: secretWriter,
		errorWriter:  errorWriter,
		destWriter:   dest,
	}
}

func (w StdoutWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w StdoutWriter) Close() error {
	if w.secretWriter != nil {
		if err := w.secretWriter.Close(); err != nil {
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
