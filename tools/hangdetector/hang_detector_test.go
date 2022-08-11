package hangdetector

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_GivenWriter_WhenTimeout_ThenHangs(t *testing.T) {
	// Given
	ticker := newMockTicker()
	detector := newHangDetector(ticker, 5, 2)
	detector.WrapOutWriter(new(bytes.Buffer))
	detector.Start()
	defer detector.Stop()

	// When
	ticker.doTicks(5)

	// Then
	assertNoTimeout(t, func(t *testing.T) { // hang detected
		<-detector.C()
	})
}

func Test_GivenWriter_WhenNoTimeout_ThenNotHangs(t *testing.T) {
	// Given
	ticker := newMockTicker()
	detector := newHangDetector(ticker, 5, 2)
	outWriter := detector.WrapOutWriter(new(bytes.Buffer))
	detector.Start()
	defer detector.Stop()

	// When
	ticker.doTicks(4)
	time.Sleep(1 * time.Second) // allow ticker channel to be drained

	_, err := outWriter.Write([]byte{0})
	require.NoError(t, err)

	ticker.doTicks(4)

	// Then
	assertTimeout(t, func(t *testing.T) { // no hang detected
		<-detector.C()
		t.Fatalf("expected no hang")
	})
}

func assertNoTimeout(t *testing.T, f func(t *testing.T)) {
	var (
		doneCh = make(chan bool)
		timer  = time.NewTimer(10 * time.Second)
	)
	defer timer.Stop()

	go func() {
		f(t)
		doneCh <- true
	}()

	select {
	case <-timer.C:
		t.Fatalf("expected no timeout")
	case <-doneCh:
		return
	}
}

func assertTimeout(t *testing.T, f func(t *testing.T)) {
	var (
		doneCh = make(chan bool)
		timer  = time.NewTimer(5 * time.Second)
	)
	defer timer.Stop()

	go func() {
		f(t)
		doneCh <- true
	}()

	select {
	case <-timer.C:
		return
	case <-doneCh:
		t.Fatalf("expected timeout")
	}
}

func Test_tickerSettings(t *testing.T) {
	tests := []struct {
		name                    string
		timeout                 time.Duration
		expectedInterval        time.Duration
		expectedTickLimit       uint64
		expectedHeartbeatAtTick uint64
	}{
		{
			name:                    "1 timeout",
			timeout:                 1 * time.Second,
			expectedInterval:        1 * time.Second,
			expectedTickLimit:       2,
			expectedHeartbeatAtTick: 1,
		},
		{
			name:                    "Small timeout",
			timeout:                 10 * time.Second,
			expectedInterval:        1 * time.Second,
			expectedTickLimit:       11,
			expectedHeartbeatAtTick: 5,
		},
		{
			name:                    "large timeout",
			timeout:                 600 * time.Second,
			expectedInterval:        10 * time.Second,
			expectedTickLimit:       61,
			expectedHeartbeatAtTick: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualInterval, actualTickLimit, actualHearthbeatAtTick := tickerSettings(tt.timeout)

			require.Equal(t, tt.expectedInterval, actualInterval)
			require.Equal(t, tt.expectedTickLimit, actualTickLimit)
			require.Equal(t, tt.expectedHeartbeatAtTick, actualHearthbeatAtTick)
		})
	}
}
