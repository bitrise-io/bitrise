package progress

import "time"

// Ticker helps with mocking time.Ticker by hiding exported struct fields
type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type ticker struct {
	wrappedTicker *time.Ticker
}

// NewTicker creates a new Ticker with the given duration
func NewTicker(d time.Duration) Ticker {
	return &ticker{
		wrappedTicker: time.NewTicker(d),
	}
}

// Chan returns the underlying ticker channel
func (t *ticker) Chan() <-chan time.Time {
	return t.wrappedTicker.C
}

// Stop stops the ticker (does not close channel)
func (t *ticker) Stop() {
	t.wrappedTicker.Stop()
}
