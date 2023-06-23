package hangdetector

import (
	"io"
)

type writer struct {
	writer           io.Writer
	writerActivityFn func()
}

func newWriter(wrappedWriter io.Writer, writerActivityFn func()) writer {
	return writer{
		writer:           wrappedWriter,
		writerActivityFn: writerActivityFn,
	}
}

func (h writer) Write(p []byte) (int, error) {
	h.writerActivityFn()
	return h.writer.Write(p)
}
