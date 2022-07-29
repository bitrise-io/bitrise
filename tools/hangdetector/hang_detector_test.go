package hangdetector

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_GivenNoWriter_WhenTimeout_ThenHangs(t *testing.T) {
	// Given
	ticker := newMockTicker()
	detector := newHangDetector(ticker, 5)
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
	detector := newHangDetector(ticker, 5)
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
		name          string
		timeout       time.Duration
		wantInterval  time.Duration
		wantTickLimit uint64
	}{
		{
			name:          "Small timeout",
			timeout:       10 * time.Second,
			wantInterval:  1 * time.Second,
			wantTickLimit: 11,
		},
		{
			name:          "large timeout",
			timeout:       600 * time.Second,
			wantInterval:  10 * time.Second,
			wantTickLimit: 61,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInterval, gotTickLimit := tickerSettings(tt.timeout)
			if gotInterval != tt.wantInterval {
				t.Errorf("tickerSettings() gotInterval = %v, want %v", gotInterval, tt.wantInterval)
			}
			if gotTickLimit != tt.wantTickLimit {
				t.Errorf("tickerSettings() gotTickLimit = %v, want %v", gotTickLimit, tt.wantTickLimit)
			}
		})
	}
}
