package hangdetector

import "time"

// Ticker helps with mocking time.Ticker by hiding exported struct fields
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type ticker struct {
	wrappedTicker *time.Ticker
}

func newTicker(d time.Duration) Ticker {
	return &ticker{
		wrappedTicker: time.NewTicker(d),
	}
}

// C returns the underlying ticker channel
func (t *ticker) C() <-chan time.Time {
	return t.wrappedTicker.C
}

// Stop stops the ticker (does not close channel)
func (t *ticker) Stop() {
	t.wrappedTicker.Stop()
}
