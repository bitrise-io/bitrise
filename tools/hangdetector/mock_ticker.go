package hangdetector

import "time"

type mockTicker struct {
	Channel chan time.Time
}

func newMockTicker() mockTicker {
	return mockTicker{
		Channel: make(chan time.Time),
	}
}

// C ...
func (t mockTicker) C() <-chan time.Time {
	return t.Channel
}

// Stop ...
func (t mockTicker) Stop() {
}

func (t mockTicker) doTicks(n int) {
	for i := 0; i < n; i++ {
		t.Channel <- time.Now()
	}
}
