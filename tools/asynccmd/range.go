package asynccmd

import (
	"bytes"
	"sort"
)

type matchRange struct {
	first int
	last  int
}

var emptyRange = matchRange{}

// allRanges returns every indexes of instance of pattern in b, or nil if pattern is not present in b.
func allRanges(b, pattern []byte) (ranges []matchRange) {
	i := 0
	for {
		sub := b[i:len(b)]
		idx := bytes.Index(sub, pattern)
		if idx == -1 {
			return
		}

		ranges = append(ranges, matchRange{first: idx + i, last: idx + i + len(pattern)})
		i += idx + 1
	}
}

// isOverlapping returns true if base and compare ranges overlaps each other, otherwise false.
func isOverlapping(base, compare matchRange) bool {
	return compare.first >= base.first && compare.first <= base.last
}

// mergeRanges returns the union of r1 and r2, r1 and r2 have to be overlapping ranges.
func mergeRanges(r1, r2 matchRange) matchRange {
	first := r1.first
	if r2.first < first {
		first = r2.first
	}
	last := r1.last
	if r2.last > last {
		last = r2.last
	}
	return matchRange{first: first, last: last}
}

// mergeAllRanges merges every overlapping ranges in r.
func mergeAllRanges(r []matchRange) []matchRange {
	ranges := append([]matchRange{}, r...)

	sort.Slice(ranges, func(i, j int) bool { return r[i].first < r[j].first })

	for i := 0; i < len(ranges); i++ {
		baseRange := ranges[i]
		if baseRange == emptyRange {
			continue
		}

		for j := i + 1; j < len(ranges); j++ {
			compareRange := ranges[j]
			if compareRange == emptyRange {
				continue
			}

			if isOverlapping(baseRange, compareRange) {
				merged := mergeRanges(baseRange, compareRange)

				baseRange = merged
				ranges[i] = merged
				ranges[j] = matchRange{}
			}
		}
	}

	var merged []matchRange
	for _, r := range ranges {
		if r != emptyRange {
			merged = append(merged, r)
		}
	}
	return merged
}
