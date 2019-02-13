package progress

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimpleProgress(t *testing.T) {
	startTime := time.Now()

	SimpleProgress(".", 500*time.Millisecond, func() {
		t.Log("- SimpleProgress [start] -")
		time.Sleep(3 * time.Second)
		t.Log("- SimpleProgress [end] -")
	})

	duration := time.Now().Sub(startTime)
	if duration >= time.Duration(4)*time.Second {
		t.Fatalf("Should take no more than 4 sec, but got: %s", duration)
	}
	if duration < time.Duration(2)*time.Second {
		t.Fatalf("Should take at least 2 sec, but got: %s", duration)
	}
}

func TestSimpleProgressE(t *testing.T) {
	t.Log("No error")
	{
		startTime := time.Now()
		actionErr := SimpleProgressE(".", 500*time.Millisecond, func() error {
			t.Log("- SimpleProgressE [start] -")
			time.Sleep(3 * time.Second)
			t.Log("- SimpleProgressE [end] -")
			return nil
		})
		require.NoError(t, actionErr)

		duration := time.Now().Sub(startTime)
		if duration >= time.Duration(4)*time.Second {
			t.Fatalf("Should take no more than 4 sec, but got: %s", duration)
		}
		if duration < time.Duration(2)*time.Second {
			t.Fatalf("Should take at least 2 sec, but got: %s", duration)
		}
	}

	t.Log("Return error")
	{
		startTime := time.Now()
		actionErr := SimpleProgressE(".", 500*time.Millisecond, func() error {
			t.Log("- SimpleProgressE [start] -")
			time.Sleep(3 * time.Second)
			t.Log("- SimpleProgressE [end] -")
			return errors.New("Test error")
		})
		require.EqualError(t, actionErr, "Test error")

		duration := time.Now().Sub(startTime)
		if duration >= time.Duration(4)*time.Second {
			t.Fatalf("Should take no more than 4 sec, but got: %s", duration)
		}
		if duration < time.Duration(2)*time.Second {
			t.Fatalf("Should take at least 2 sec, but got: %s", duration)
		}
	}
}
