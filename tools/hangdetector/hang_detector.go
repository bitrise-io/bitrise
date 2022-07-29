package hangdetector

import (
	"io"
	"sync/atomic"
	"time"
)

// HangDetector ...
type HangDetector interface {
	Start()
	Stop()
	C() <-chan bool
	WrapOutWriter(writer io.Writer) io.Writer
	WrapErrWriter(writer io.Writer) io.Writer
}

type hangDetector struct {
	ticker        Ticker
	ticks         uint64
	tickLimit     uint64
	notificationC chan bool
	stopC         chan bool
}

// NewDefaultHangDetector ...
func NewDefaultHangDetector(timeout time.Duration) HangDetector {
	const tickerInterval = time.Second * 30
	maxIntervals := uint64(timeout / tickerInterval)

	return newHangDetector(NewTicker(tickerInterval), maxIntervals)
}

func newHangDetector(ticker Ticker, maxIntervals uint64) HangDetector {
	detector := hangDetector{
		ticker:        ticker,
		tickLimit:     maxIntervals,
		notificationC: make(chan bool, 1),
		stopC:         make(chan bool, 1),
	}

	return &detector
}

// Start ...
func (h *hangDetector) Start() {
	go func() {
		for {
			select {
			case <-h.ticker.C():
				{
					count := atomic.AddUint64(&h.ticks, 1)
					if count >= h.tickLimit {
						h.notificationC <- true
						return
					}
				}
			case <-h.stopC:
				return
			}
		}
	}()
}

// Stop ...
func (h *hangDetector) Stop() {
	h.ticker.Stop()
	h.stopC <- true
}

// C ...
func (h *hangDetector) C() <-chan bool {
	return h.notificationC
}

// WrapOutWriter ...
func (h *hangDetector) WrapOutWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.onWriterActivity)

	return hangWriter
}

// WrapErrWriter ...
func (h *hangDetector) WrapErrWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.onWriterActivity)

	return hangWriter
}

func (h *hangDetector) onWriterActivity() {
	atomic.StoreUint64(&h.ticks, 0)
}
