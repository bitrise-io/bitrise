package asynccmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_mergeAllRanges(t *testing.T) {
	t.Log("merges overlapping ranges")
	{
		ranges := []matchRange{
			{first: 0, last: 2},
			{first: 1, last: 3},
		}

		merged := mergeAllRanges(ranges)
		require.Equal(t, []matchRange{
			{first: 0, last: 3},
		}, merged)
	}

	t.Log("does not merge distinct ranges")
	{
		ranges := []matchRange{
			{first: 0, last: 2},
			{first: 3, last: 5},
		}

		merged := mergeAllRanges(ranges)
		require.Equal(t, []matchRange{
			{first: 0, last: 2},
			{first: 3, last: 5},
		}, merged)
	}

	t.Log("returns the wider range")
	{
		ranges := []matchRange{
			{first: 0, last: 2},
			{first: 1, last: 2},
		}

		merged := mergeAllRanges(ranges)
		require.Equal(t, []matchRange{
			{first: 0, last: 2},
		}, merged)
	}

	t.Log("complex test")
	{
		ranges := []matchRange{
			{first: 11, last: 15},
			{first: 0, last: 2},
			{first: 11, last: 13},
			{first: 2, last: 4},
			{first: 6, last: 9},
			{first: 5, last: 10},
		}

		merged := mergeAllRanges(ranges)
		require.Equal(t, []matchRange{
			{first: 0, last: 4},
			{first: 11, last: 15},
			{first: 5, last: 10},
		}, merged)
	}
}
