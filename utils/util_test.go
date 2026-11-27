package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeToFormattedSeconds(t *testing.T) {
	t.Log("formatted print rounds")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(999) * time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "1.00 sec", timeStr)
	}

	t.Log("sec < 1.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(111) * time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "0.11 sec", timeStr)
	}

	t.Log("sec < 60.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(59)*time.Second + time.Duration(111)*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "59.11 sec", timeStr)
	}

	t.Log("min < 60")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(59)*time.Minute + time.Duration(6660)*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "59.1 min", timeStr)
	}

	t.Log("hour < 10")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(9)*time.Hour + time.Duration(399600)*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "9.1 hour", timeStr)
	}

	t.Log("hour < 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(999)*time.Hour + time.Duration(399600)*time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, "999 hour", timeStr)
	}

	t.Log("hour >= 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(time.Duration(1000) * time.Hour)
		require.EqualError(t, err, "time (1000.000000 hour) greater than max allowed (999 hour)")
		require.Equal(t, "", timeStr)
	}
}

func TestIsExecutable(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, mode os.FileMode) string {
		pth := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(pth, []byte("#!/bin/sh\n"), mode))
		require.NoError(t, os.Chmod(pth, mode))
		return pth
	}

	require.True(t, IsExecutable(write("owner", 0o755)))
	require.False(t, IsExecutable(write("none", 0o644)))
	require.False(t, IsExecutable(dir), "a directory")
	if os.Geteuid() != 0 {
		// The mode bits allow group and other, but not the owner running the test.
		require.False(t, IsExecutable(write("others-only", 0o055)))
	}
}
