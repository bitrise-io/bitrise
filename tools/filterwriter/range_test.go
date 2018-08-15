package filterwriter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllRanges(t *testing.T) {
	{
		ranges := allRanges([]byte("test"), []byte("t"))
		require.Equal(t, []matchRange{{first: 0, last: 1}, {first: 3, last: 4}}, ranges)
	}

	{
		ranges := allRanges([]byte("test rangetest"), []byte("test"))
		require.Equal(t, []matchRange{{first: 0, last: 4}, {first: 10, last: 14}}, ranges)
	}

	{
		ranges := allRanges([]byte("\n"), []byte("\n"))
		require.Equal(t, []matchRange{{first: 0, last: 1}}, ranges)
	}

	{
		ranges := allRanges([]byte("test\n"), []byte("\n"))
		require.Equal(t, []matchRange{{first: 4, last: 5}}, ranges)
	}

	{
		ranges := allRanges([]byte("\n\ntest\n"), []byte("\n"))
		require.Equal(t, []matchRange{{first: 0, last: 1}, {first: 1, last: 2}, {first: 6, last: 7}}, ranges)
	}

	{
		ranges := allRanges([]byte("\n\ntest\n"), []byte("test\n"))
		require.Equal(t, []matchRange{{first: 2, last: 7}}, ranges)
	}
}

func TestMergeAllRanges(t *testing.T) {
	var testCases = []struct {
		name   string
		ranges []matchRange
		want   []matchRange
	}{
		{
			name:   "merges overlapping ranges",
			ranges: []matchRange{{0, 2}, {1, 3}},
			want:   []matchRange{{0, 3}},
		},
		{
			name:   "does not merge distinct ranges",
			ranges: []matchRange{{0, 2}, {3, 5}},
			want:   []matchRange{{0, 2}, {3, 5}},
		},
		{
			name:   "returns the wider range",
			ranges: []matchRange{{0, 2}, {1, 2}},
			want:   []matchRange{{0, 2}},
		},
		{
			name:   "complex test",
			ranges: []matchRange{{11, 15}, {0, 2}, {11, 13}, {2, 4}, {6, 9}, {5, 10}},
			want:   []matchRange{{0, 4}, {5, 10}, {11, 15}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mergeAllRanges(tc.ranges); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}

}
