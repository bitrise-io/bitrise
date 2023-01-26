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
	RunError() error
}

type writer struct {
	writer io.Writer
}

func NewWriter(isSecretFiltering bool, secrets []string, stepUUID string) Writer {
	opts := log.GetGlobalLoggerOpts()
	opts.Producer = log.Step
	opts.ProducerID = stepUUID
	logWriter := logwriter.NewLogWriter(log.NewLogger(opts))

	var w io.Writer

	errorFinder := errorfinder.NewErrorFinder()
	var fw *filterwriter.Writer

	if !isSecretFiltering {
		w = errorFinder.WrapWriter(logWriter)
	} else {
		fw = filterwriter.New(secrets, logWriter)
		w = errorFinder.WrapWriter(fw)
	}

	return writer{
		writer: w,
	}
}

func (w writer) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w writer) Flush() (int, error) {
	return 0, nil
}

func (w writer) RunError() error {
	// todo: unwarp and return error from the errorfinder writer
	return nil
}
