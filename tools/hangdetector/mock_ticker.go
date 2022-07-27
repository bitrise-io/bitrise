package hangdetector

import "time"

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
