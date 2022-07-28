package hangdetector

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

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
		ticker:        ticker,
		tickLimit:     maxIntervals,
		notificationC: make(chan bool, 1),
		stopC:         make(chan bool, 1),
	}

	return &detector
}

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
				log.Infof("ticker exited")
				return
			}
		}
	}()
}

func (h *hangDetector) Stop() {
	h.ticker.Stop()
	h.stopC <- true
}

func (h *hangDetector) C() <-chan bool {
	return h.notificationC
}

func (h *hangDetector) WrapOutWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.onWriterActivity)
	h.outWriter = hangWriter

	return hangWriter
}

func (h *hangDetector) WrapErrWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.onWriterActivity)
	h.errWriter = hangWriter

	return hangWriter
}

func (h *hangDetector) onWriterActivity() {
	atomic.StoreUint64(&h.ticks, 0)
}
