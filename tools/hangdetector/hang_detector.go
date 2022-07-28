package hangdetector

import (
	"io"
	"sync"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

type HangDetector interface {
	Start()
	Stop()
	C() chan bool
	WrapOutWriter(writer io.Writer) io.Writer
	WrapErrWriter(writer io.Writer) io.Writer
}

type hangDetector struct {
	ticker           Ticker
	ticks            uint64
	tickLimit        uint64
	notification     chan bool
	writerActivityFn func()
	mutex            sync.Mutex
}

func NewDefaultHangDetector(timeout time.Duration) HangDetector {
	const tickerInterval = time.Second * 1
	tickLimit := uint64(timeout / tickerInterval)

	return newHangDetector(NewTicker(tickerInterval), tickLimit)
}

func newHangDetector(ticker Ticker, tickLimit uint64) HangDetector {
	log.Warnf("tick limit: %d", tickLimit)
	detector := hangDetector{
		ticker:       ticker,
		ticks:        0,
		tickLimit:    tickLimit,
		notification: make(chan bool, 1),
		mutex:        sync.Mutex{},
	}
	detector.writerActivityFn = func() {
		detector.mutex.Lock()
		defer detector.mutex.Unlock()
		detector.ticks = 0
	}

	return &detector
}

func (h *hangDetector) Start() {
	h.checkHang()
}

func (h *hangDetector) Stop() {
	h.ticker.Stop()
}

func (h *hangDetector) WrapOutWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.writerActivityFn)
	return hangWriter
}

func (h *hangDetector) WrapErrWriter(writer io.Writer) io.Writer {
	hangWriter := newWriter(writer, h.writerActivityFn)
	return hangWriter
}

func (h *hangDetector) C() chan bool {
	return h.notification
}

func (h *hangDetector) checkHang() {
	go func() {
		for range h.ticker.C() {
			h.mutex.Lock()
			h.ticks++
			log.Warnf("tick #%d", h.ticks)
			tickLimitReached := h.ticks >= h.tickLimit
			h.mutex.Unlock()

			if tickLimitReached {
				log.Warnf("tick limit reached")
				h.notification <- true
				return
			}
		}
	}()
}
