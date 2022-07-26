package hangdetector

import (
	"io"
	"sync/atomic"
)

type HangDetectorWriter interface {
	io.Writer
}

type hangDetectorWriter struct {
	writer io.Writer
	count  *uint64
}

func NewHangDetectorWriter(writer io.Writer, count *uint64) HangDetectorWriter {
	return hangDetectorWriter{
		writer: writer,
		count:  count,
	}
}

func (h hangDetectorWriter) Write(p []byte) (int, error) {
	atomic.StoreUint64(h.count, 0)
	return h.writer.Write(p)
}
