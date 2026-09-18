package purr

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurr_StaticToBuffer(t *testing.T) {
	// *bytes.Buffer is never a TTY → the static path runs even without
	// --once. Confirms that a non-TTY destination never animates and
	// always emits ANSI-free output.
	var buf bytes.Buffer
	require.NoError(t, runPurr(t.Context(), nil, &buf, false, time.Second, time.Millisecond))
	out := buf.String()
	assert.NotContains(t, out, "\x1b[")
	assert.Contains(t, out, purrMessage)
	// Cat ear lines + message + leading blank = exactly 12 lines.
	assert.Equal(t, 12, strings.Count(out, "\n"))
}

func TestPurr_OnceDoesntAnimate(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, runPurr(t.Context(), nil, &buf, true, time.Hour, time.Microsecond))
	// With once=true the function returns immediately after one paint; it
	// must not block on the ticker even with a microsecond interval.
	assert.Contains(t, buf.String(), "Purr Request")
}

func TestPurr_StaticMessageHasNoANSIOnBuffer(t *testing.T) {
	// Even with the rainbow effect, output to *bytes.Buffer must remain
	// ANSI-free — that's the contract that keeps log files / pipes / JSON
	// output clean.
	var buf bytes.Buffer
	require.NoError(t, runPurr(t.Context(), nil, &buf, true, time.Second, time.Millisecond))
	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestCRLFWriter_TranslatesNewlines(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello", "hello"},
		{"line1\nline2", "line1\r\nline2"},
		{"a\nb\nc", "a\r\nb\r\nc"},
		{"\n\n", "\r\n\r\n"},
		// Already-CRLF should NOT be doubled — \r passes through, then \n becomes \r\n.
		{"a\r\nb", "a\r\r\nb"},
		// Cursor escape codes don't contain \n, must pass through untouched.
		{"\x1b[5Fhello", "\x1b[5Fhello"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		w := &crlfWriter{w: &buf}
		n, err := w.Write([]byte(c.in))
		require.NoError(t, err)
		assert.Equal(t, len(c.in), n, "Write(%q) return value per io.Writer contract", c.in)
		assert.Equal(t, c.want, buf.String())
	}
}

func TestPurrFrames_AllSameShape(t *testing.T) {
	// Animation looks ugly if frames have different heights or the same
	// content, so guard both invariants.
	require.GreaterOrEqual(t, len(purrFrames), 2)
	height := strings.Count(purrFrames[0], "\n")
	for i, f := range purrFrames {
		assert.Equal(t, height, strings.Count(f, "\n"), "frame[%d] height", i)
	}
	seen := make(map[string]bool)
	distinct := 0
	for _, f := range purrFrames {
		if !seen[f] {
			seen[f] = true
			distinct++
		}
	}
	assert.GreaterOrEqual(t, distinct, 2, "frames are all identical — animation has no motion")
}

func TestNewCmd_Flags(t *testing.T) {
	cmd := NewCmd()
	assert.True(t, cmd.Hidden)

	once, err := cmd.Flags().GetBool("once")
	require.NoError(t, err)
	assert.False(t, once)

	duration, err := cmd.Flags().GetDuration("duration")
	require.NoError(t, err)
	assert.Equal(t, 8*time.Second, duration)

	interval, err := cmd.Flags().GetDuration("interval")
	require.NoError(t, err)
	assert.Equal(t, 250*time.Millisecond, interval)
}

func TestNewCmd_RunEUsesStaticPathOnBuffer(t *testing.T) {
	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	require.NoError(t, cmd.Flags().Set("once", "true"))
	require.NoError(t, cmd.RunE(cmd, nil))
	assert.Contains(t, out.String(), purrMessage)
}
