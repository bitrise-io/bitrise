package hangdetector

import (
	"io"
	"sync/atomic"
	"time"
)

type HangDetector interface {
	Start()
	Stop()
	C() chan bool
	WrapOutWriter(writer io.Writer) io.Writer
	WrapErrWriter(writer io.Writer) io.Writer
}

type hangDetector struct {
	ticker       Ticker
	ticks        uint64
	tickLimit    uint64
	notification chan bool

	outWriter writer
	errWriter writer
}

func NewDefaultHangDetector(timeout time.Duration) HangDetector {
	const tickerInterval = time.Second * 30
	maxIntervals := uint64(timeout / tickerInterval)

	return newHangDetector(NewTicker(tickerInterval), maxIntervals)
}

func newHangDetector(ticker Ticker, maxIntervals uint64) HangDetector {
	detector := hangDetector{
		ticker:       ticker,
		tickLimit:    maxIntervals,
		notification: make(chan bool, 1),
	}
	detector.checkHang()

	return &detector
}

func (h *hangDetector) Start() {
	h.checkHang()
}

func (h *hangDetector) Stop() {
	h.ticker.Stop()
}

func (h *hangDetector) C() chan bool {
	return h.notification
}

func (h *hangDetector) WrapOutWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, &h.ticks)
	h.outWriter = hangWriter

	return hangWriter
}

func (h *hangDetector) WrapErrWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, &h.ticks)
	h.errWriter = hangWriter

	return hangWriter
}

func (h *hangDetector) checkHang() {
	go func() {
		for range h.ticker.C() {
			count := atomic.AddUint64(&h.ticks, 1)
			if count >= h.tickLimit {
				h.notification <- true
			}
		}
	}()
}
