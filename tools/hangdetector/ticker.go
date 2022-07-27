package hangdetector

import "time"

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type ticker struct {
	ticker *time.Ticker
}

func NewTicker(d time.Duration) Ticker {
	return &ticker{
		ticker: time.NewTicker(d),
	}
}

func (t *ticker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *ticker) Stop() {
	t.ticker.Stop()
}
