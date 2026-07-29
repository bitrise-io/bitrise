// Package stringutil holds small string helpers shared across the internal
// packages. Depends on nothing but the standard library, so any of them can
// import it.
package stringutil

import "unicode/utf8"

// Truncate shortens s to at most limit bytes, marking the cut with an ellipsis.
// The cut lands on a rune boundary, so the result stays valid UTF-8 — callers
// truncate HTTP response bodies for error messages, which are usually UTF-8.
func Truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	if limit <= 0 {
		return "…"
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}
