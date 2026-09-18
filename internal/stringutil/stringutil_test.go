package stringutil

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, "short", Truncate("short", 10), "a string within the limit is returned unchanged")
	assert.Equal(t, "exact", Truncate("exact", 5), "a string exactly at the limit is not cut")
	assert.Equal(t, "abcde…", Truncate("abcdefgh", 5))
	assert.Equal(t, "", Truncate("", 5))
	assert.Equal(t, "…", Truncate("abc", 0))
}

// TestTruncate_CutsOnRuneBoundary is the reason this is shared rather than
// re-implemented per package: a byte-wise cut through a multi-byte rune emits
// invalid UTF-8 into the error message the snippet ends up in.
func TestTruncate_CutsOnRuneBoundary(t *testing.T) {
	// Each ° is 2 bytes, so a limit of 3 lands mid-rune.
	got := Truncate("°°°", 3)
	assert.True(t, utf8.ValidString(got), "result should stay valid UTF-8, got %q", got)
	assert.Equal(t, "°…", got)

	// Every cut position of a 3-byte-rune string must stay valid.
	s := strings.Repeat("→", 4)
	for limit := 1; limit < len(s); limit++ {
		got := Truncate(s, limit)
		assert.True(t, utf8.ValidString(got), "limit %d produced invalid UTF-8: %q", limit, got)
	}
}
