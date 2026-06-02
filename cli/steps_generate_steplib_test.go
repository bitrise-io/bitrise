package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildSpecgenOpts(t *testing.T) {
	t.Run("commit-sha passes through verbatim", func(t *testing.T) {
		opts, err := buildSpecgenOpts("deadbeef", "")
		require.NoError(t, err)
		require.Equal(t, "deadbeef", opts.SteplibCommitSHA)
		require.True(t, opts.GeneratedAt.IsZero(), "GeneratedAt should stay zero so specgen defaults it")
	})

	t.Run("both fields empty yields zero-value Options", func(t *testing.T) {
		opts, err := buildSpecgenOpts("", "")
		require.NoError(t, err)
		require.Empty(t, opts.SteplibCommitSHA)
		require.True(t, opts.GeneratedAt.IsZero())
	})

	t.Run("valid RFC3339 timestamp populates GeneratedAt", func(t *testing.T) {
		opts, err := buildSpecgenOpts("", "2026-01-15T10:30:00Z")
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC), opts.GeneratedAt)
	})

	t.Run("invalid timestamp returns error mentioning --timestamp", func(t *testing.T) {
		_, err := buildSpecgenOpts("", "not-a-timestamp")
		require.Error(t, err)
		require.Contains(t, err.Error(), "--timestamp")
	})
}
