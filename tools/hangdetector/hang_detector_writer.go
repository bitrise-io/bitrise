package hangdetector

import (
	"io"
	"sync/atomic"
)

type writer struct {
	writer io.Writer
	count  *uint64
}

func newWriter(wrappedWriter io.Writer, count *uint64) io.Writer {
	return writer{
		writer: wrappedWriter,
		count:  count,
	}
}

func (h writer) Write(p []byte) (int, error) {
	atomic.StoreUint64(h.count, 0)
	return h.writer.Write(p)
}
