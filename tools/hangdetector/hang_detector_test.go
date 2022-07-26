package hangdetector

import (
	"testing"
)

func Test_GivenNoWriter_WhenTimeout_ThenHangs_(t *testing.T) {
	ticker := NewMockTicker()
	detector := newHangDetector(ticker, 5)

	ticker.DoTicks(5)

	<-detector.C()
}
