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
	return ticker{
		ticker: time.NewTicker(d),
	}
}

func (t ticker) C() <-chan time.Time {
	return t.ticker.C
}

func (t ticker) Stop() {
	t.ticker.Stop()
}

type MockTicker struct {
	Channel chan time.Time
}

func NewMockTicker() MockTicker {
	return MockTicker{
		Channel: make(chan time.Time),
	}
}

func (t MockTicker) C() <-chan time.Time {
	return t.Channel
}

func (t MockTicker) Stop() {
}

func (t MockTicker) DoTicks(n int) {
	for i := 0; i < n; i++ {
		t.Channel <- time.Now()
	}
}
