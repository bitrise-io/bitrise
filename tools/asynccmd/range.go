package asynccmd

import (
	"bytes"
	"sort"
)

type matchRange struct{ first, last int }

// allRanges returns every indexes of instance of pattern in b, or nil if pattern is not present in b.
func allRanges(b, pattern []byte) (ranges []matchRange) {
	for i, idx := 0, bytes.Index(b, pattern); idx != -1; idx = bytes.Index(b[i:], pattern) {
		ranges = append(ranges, matchRange{idx + i, idx + i + len(pattern)})
		i += idx + 1
	}
	return
}

// mergeAllRanges merges every overlapping ranges in r.
func mergeAllRanges(r []matchRange) []matchRange {
	sort.Slice(r, func(i, j int) bool { return r[i].first < r[j].first })
	for i := 0; i < len(r)-1; i++ {
		for i+1 < len(r) && r[i+1].first <= r[i].last {
			if r[i+1].last > r[i].last {
				r[i].last = r[i+1].last
			}
			r = append(r[:i+1], r[i+2:]...)
		}
	}
	return r
}
