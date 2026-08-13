package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
)

// The TUI used to print each log line with tea.Batch, which bubbletea runs
// concurrently — so multi-line chunks and back-to-back chunks scrambled on
// screen even though Watch delivered them in order. These tests pin the
// single-flight serialization that replaced it: at most one print is in
// flight, and lines arriving during a print buffer until it completes.

func TestWaitModel_SerializesLogOutputSingleFlight(t *testing.T) {
	m := newWaitModel(internalbuild.Build{BuildNumber: 1})

	// First chunk with two complete lines starts exactly one print and
	// drains the lines into that in-flight block.
	next, cmd := m.Update(logChunkMsg("alpha\nbeta\n"))
	wm := next.(waitModel)
	assert.True(t, wm.printing, "expected printing=true after first chunk")
	assert.Empty(t, wm.pending, "expected pending drained into print block")
	require.NotNil(t, cmd, "expected a print command")

	// A second chunk arriving while the print is in flight must buffer in
	// pending and must NOT issue a second, racing print command.
	next, cmd = wm.Update(logChunkMsg("gamma\n"))
	wm = next.(waitModel)
	assert.True(t, wm.printing, "expected still printing")
	assert.Equal(t, []string{"gamma"}, wm.pending)
	assert.Nil(t, cmd, "expected no new command while a print is in flight")

	// When the in-flight print finishes, the buffered line flushes.
	next, cmd = wm.Update(printDoneMsg{})
	wm = next.(waitModel)
	assert.True(t, wm.printing, "expected a new print for the buffered line")
	assert.Empty(t, wm.pending, "expected pending drained")
	require.NotNil(t, cmd, "expected print command for buffered line")

	// With nothing left buffered, the next done returns to idle.
	next, _ = wm.Update(printDoneMsg{})
	wm = next.(waitModel)
	assert.False(t, wm.printing, "expected printing=false when nothing buffered")
}

func TestWaitModel_HoldsPartialLineUntilNewline(t *testing.T) {
	m := newWaitModel(internalbuild.Build{})

	// A chunk without a trailing newline isn't a complete line yet, so
	// nothing prints.
	next, cmd := m.Update(logChunkMsg("partial"))
	wm := next.(waitModel)
	assert.False(t, wm.printing)
	assert.Nil(t, cmd, "expected no print for an incomplete line")
	assert.Equal(t, "partial", wm.leftover)

	// Completing the line triggers the print.
	next, cmd = wm.Update(logChunkMsg(" line\n"))
	wm = next.(waitModel)
	assert.True(t, wm.printing, "expected a print once the newline arrives")
	assert.NotNil(t, cmd)
	assert.Empty(t, wm.leftover, "expected leftover consumed")
}

func TestWaitModel_FinalFlushEmitsTrailingPartialAndQuits(t *testing.T) {
	m := newWaitModel(internalbuild.Build{})

	// A trailing partial line (a build's last line often has no newline).
	next, _ := m.Update(logChunkMsg("ExitCode: 0"))
	wm := next.(waitModel)

	next, cmd := wm.Update(watchDoneMsg{build: internalbuild.Build{Status: "success"}})
	wm = next.(waitModel)
	assert.True(t, wm.finished)
	assert.Empty(t, wm.pending, "expected pending flushed at exit")
	assert.Empty(t, wm.leftover, "expected leftover flushed at exit")
	assert.NotNil(t, cmd, "expected a flush+quit command")
}

func TestWaitModel_WaitsForInFlightPrintBeforeQuitting(t *testing.T) {
	m := newWaitModel(internalbuild.Build{})

	// Start a print.
	next, _ := m.Update(logChunkMsg("line\n"))
	wm := next.(waitModel)
	require.True(t, wm.printing)

	// Build finishes while that print is still in flight: the model must
	// mark itself quitting but NOT issue the quit yet (that would race the
	// in-flight print and could drop or reorder the final lines).
	next, cmd := wm.Update(watchDoneMsg{})
	wm = next.(waitModel)
	assert.True(t, wm.quitting)
	assert.Nil(t, cmd, "expected nil command (waiting for the in-flight print)")

	// Once the in-flight print signals done, the final flush + quit runs.
	_, cmd = wm.Update(printDoneMsg{})
	assert.NotNil(t, cmd, "expected flush+quit after the in-flight print completes")
}
