//go:build steplib_e2e

package steplibe2e

import "sort"

// diffLogs returns the log lines (by level+normalized-message) present in a but
// not b (aOnly) and present in b but not a (bOnly). Duplicates within a side are
// collapsed.
func diffLogs(a, b []logLine) (aOnly, bOnly []logLine) {
	aSet, bSet := keySet(a), keySet(b)
	aOnly = uniqueMissing(a, bSet)
	bOnly = uniqueMissing(b, aSet)
	return aOnly, bOnly
}

func keySet(lines []logLine) map[string]bool {
	m := make(map[string]bool, len(lines))
	for _, l := range lines {
		m[l.key()] = true
	}
	return m
}

func uniqueMissing(lines []logLine, other map[string]bool) []logLine {
	var out []logLine
	seen := map[string]bool{}
	for _, l := range lines {
		if other[l.key()] || seen[l.key()] {
			continue
		}
		seen[l.key()] = true
		out = append(out, l)
	}
	return out
}

// pairDiff is the log diff between v1-source and one v2 variant for a step+version.
type pairDiff struct {
	v2Variant string  // "v2-source" or "v2-precompiled"
	v2Status  string  // "OK" / "FAILED: ..." for that v2 cell
	v1Only    []logLine // logged by v1-source, not by the v2 variant
	v2Only    []logLine // logged by the v2 variant, not by v1-source
}

// comparison holds, for one step+version, the v1-source baseline status and the
// diffs of each v2 variant against it.
type comparison struct {
	step         string
	versionLabel string
	versionRef   string
	v1Status     string
	pairs        []pairDiff
}

// tally counts how often a given (level|message) line shows up as a divergence
// across the whole matrix, for the roll-up section.
type tally struct {
	line  logLine
	count int
}

func rollup(lines [][]logLine) []tally {
	counts := map[string]*tally{}
	for _, group := range lines {
		for _, l := range group {
			t, ok := counts[l.key()]
			if !ok {
				t = &tally{line: l}
				counts[l.key()] = t
			}
			t.count++
		}
	}
	out := make([]tally, 0, len(counts))
	for _, t := range counts {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].line.key() < out[j].line.key()
	})
	return out
}
