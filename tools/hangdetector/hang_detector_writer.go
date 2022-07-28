package hangdetector

import "io"

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

func (w writer) Write(p []byte) (int, error) {
	w.writerActivityFn()
	return w.writer.Write(p)
}
