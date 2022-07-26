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
	ticker        time.Ticker
	intervalCount uint64
	maxIntervals  uint64
	notification  chan bool

	writers []HangDetectorWriter
}

func NewHangDetector(timeout time.Duration) HangDetector {
	tickerInterval := time.Second * 30
	maxIntervals := uint64(timeout / tickerInterval)

	detector := hangDetector{
		ticker:       *time.NewTicker(tickerInterval),
		maxIntervals: maxIntervals,
		notification: make(chan bool),
	}
	detector.checkHang()

	return &detector
}

func (h *hangDetector) WrapWriter(writer io.Writer) io.Writer {
	hangWriter := NewHangDetectorWriter(writer, &h.intervalCount)
	h.writers = append(h.writers, hangWriter)

	return hangWriter
}

func (h *hangDetector) C() chan bool {
	return h.notification
}

func (h hangDetector) checkHang() {
	go func() {
		for range h.ticker.C {
			count := atomic.AddUint64(&h.intervalCount, 1)
			if count >= h.maxIntervals {
				h.notification <- true
			}
		}
	}()
}
