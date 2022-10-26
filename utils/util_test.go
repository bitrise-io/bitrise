package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func secToDuration(sec float64) time.Duration {
	return time.Duration(sec * 1e9)
}

func minToDuration(min float64) time.Duration {
	return secToDuration(min * 60)
}

func hourToDuration(hour float64) time.Duration {
	return minToDuration(hour * 60)
}

func TestTimeToFormattedSeconds(t *testing.T) {
	t.Log("formatted print rounds")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(0.999))
		require.NoError(t, err)
		require.Equal(t, "1.00 sec", timeStr)
	}

	t.Log("sec < 1.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(0.111))
		require.NoError(t, err)
		require.Equal(t, "0.11 sec", timeStr)
	}

	t.Log("sec < 60.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(59.111))
		require.NoError(t, err)
		require.Equal(t, "59.11 sec", timeStr)
	}

	t.Log("min < 60")
	{
		timeStr, err := FormattedSecondsToMax8Chars(minToDuration(59.111))
		require.NoError(t, err)
		require.Equal(t, "59.1 min", timeStr)
	}

	t.Log("hour < 10")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(9.111))
		require.NoError(t, err)
		require.Equal(t, "9.1 hour", timeStr)
	}

	t.Log("hour < 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(999.111))
		require.NoError(t, err)
		require.Equal(t, "999 hour", timeStr)
	}

	t.Log("hour >= 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(1000))
		require.EqualError(t, err, "time (1000.000000 hour) greater than max allowed (999 hour)")
		require.Equal(t, "", timeStr)
	}
}
