package hangdetector

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/bitrise-io/go-utils/log"
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
	ticker                      Ticker
	ticks                       uint64
	tickLimit, hearthbeatAtTick uint64
	notificationC, stopC        chan bool

	outWriter io.Writer
}

// NewDefaultHangDetector ...
func NewDefaultHangDetector(timeout time.Duration) HangDetector {
	tickerInterval, tickLimit, hearthbeatAtTick := tickerSettings(timeout)

	return newHangDetector(newTicker(tickerInterval), tickLimit, hearthbeatAtTick)
}

func newHangDetector(ticker Ticker, tickLimit, hearthbeatAtTick uint64) HangDetector {
	detector := hangDetector{
		ticker:           ticker,
		tickLimit:        tickLimit,
		hearthbeatAtTick: hearthbeatAtTick,
		notificationC:    make(chan bool, 1),
		stopC:            make(chan bool, 1),
	}

	return &detector
}

// Start ...
func (h *hangDetector) Start() {
	if h.outWriter == nil {
		panic("Output is not set")
	}

	go func() {
		for {
			select {
			case <-h.ticker.C():
				{
					count := atomic.AddUint64(&h.ticks, 1)
					if count == h.hearthbeatAtTick {
						log.TPrintf("No output received for a while. Bitrise CLI is still active.")
					}
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
	h.outWriter = newWriter(writer, h.onWriterActivity)

	return h.outWriter
}

// WrapErrWriter ...
func (h *hangDetector) WrapErrWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.onWriterActivity)

	return hangWriter
}

func (h *hangDetector) onWriterActivity() {
	atomic.StoreUint64(&h.ticks, 0)
}

func tickerSettings(timeout time.Duration) (interval time.Duration, tickLimit, hearthbeatLimit uint64) {
	// For longer timeouts using a longer ticker interval.
	interval = 10 * time.Second
	if timeout < 5*time.Minute {
		interval = time.Second
	}

	tickLimit = uint64(timeout/interval) + 1
	hearthbeatLimit = tickLimit / 2

	return
}
