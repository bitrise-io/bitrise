package hangdetector

import (
	"io"
	"sync/atomic"
	"time"
)

type HangDetector interface {
	WrapWriter(writer io.Writer) io.Writer
	C() chan bool
}

type hangDetector struct {
	ticker               Ticker
	elapsedIntervalCount uint64
	maxIntervals         uint64
	notification         chan bool

	writers []io.Writer
}

func NewDefaultHangDetector(timeout time.Duration) HangDetector {
	const tickerInterval = time.Second * 30
	maxIntervals := uint64(timeout / tickerInterval)

	return newHangDetector(NewTicker(tickerInterval), maxIntervals)
}

func newHangDetector(ticker Ticker, maxIntervals uint64) HangDetector {
	detector := hangDetector{
		ticker:       ticker,
		maxIntervals: maxIntervals,
		notification: make(chan bool, 1),
	}
	detector.checkHang()

	return &detector
}

func (h *hangDetector) WrapWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, &h.elapsedIntervalCount)
	h.writers = append(h.writers, hangWriter)

	return hangWriter
}

func (h *hangDetector) C() chan bool {
	return h.notification
}

func (h *hangDetector) checkHang() {
	go func() {
		for range h.ticker.C() {
			count := atomic.AddUint64(&h.elapsedIntervalCount, 1)
			if count >= h.maxIntervals {
				h.notification <- true
			}
		}
	}()
}
