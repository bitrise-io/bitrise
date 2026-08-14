package build

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Watch_DeltaStreaming(t *testing.T) {
	// Three log polls: chunk1 (ts1), chunk2 (ts2), chunk3 (empty next ts).
	// Build status returns in-progress until the log stream ends naturally,
	// then the final-flush call and a View call complete the sequence.
	var logCalls, buildCalls atomic.Int32
	var logTimestamps []string

	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			n := int(logCalls.Add(1))
			logTimestamps = append(logTimestamps, r.URL.Query().Get("after_timestamp"))
			switch n {
			case 1:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"chunk1\n","position":0}],"next_after_timestamp":"ts1"}`))
			case 2:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"chunk2\n","position":1}],"next_after_timestamp":"ts2"}`))
			case 3:
				// Empty next_after_timestamp: loop exits after this poll.
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"chunk3\n","position":2}]}`))
			default: // final flush
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"final\n","position":3}]}`))
			}
		case "/apps/my-app/builds/b-1":
			n := int(buildCalls.Add(1))
			// Stay in-progress while the log stream is active so we can observe
			// all chunks; return success on the final View call.
			if n <= 2 {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":1,"triggered_workflow":"primary","branch":"main"}}`))
			}
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	build, err := svc.Watch(context.Background(), "my-app", "b-1", &buf, time.Millisecond)
	require.NoError(t, err)

	got := buf.String()
	for _, want := range []string{"chunk1", "chunk2", "chunk3", "final"} {
		assert.Contains(t, got, want)
	}

	// Verify after_timestamp progression for log calls: "", ts1, ts2, ts2 (final flush).
	wantTimestamps := []string{"", "ts1", "ts2", "ts2"}
	require.GreaterOrEqual(t, len(logTimestamps), len(wantTimestamps))
	for i, want := range wantTimestamps {
		assert.Equal(t, want, logTimestamps[i], "log call %d", i+1)
	}

	assert.Equal(t, "success", build.Status)
}

func TestService_Watch_CrossPollGapFillsInLaterFlushesInOrder(t *testing.T) {
	// Realistic cross-poll out-of-order: poll 1 returns positions 0, 1, 3
	// (position 2 is delayed by the server-side parallel producer); poll 2
	// delivers position 2. Without cross-poll buffering the user would see 3
	// emitted before 2; with the buffer we emit 0, 1, hold 3, then emit 2
	// and 3 in order once 2 arrives.
	var logCalls, buildCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			n := int(logCalls.Add(1))
			switch n {
			case 1:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"A\n","position":0},{"chunk":"B\n","position":1},{"chunk":"D\n","position":3}],"next_after_timestamp":"ts1"}`))
			case 2:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"C\n","position":2}],"next_after_timestamp":"ts2"}`))
			default:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[]}`))
			}
		case "/apps/my-app/builds/b-1":
			n := int(buildCalls.Add(1))
			if n <= 2 {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":1,"triggered_workflow":"primary","branch":"main"}}`))
			}
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	_, err := svc.Watch(context.Background(), "my-app", "b-1", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "A\nB\nC\nD\n", buf.String(), "chunk D must be held until C arrives")
}

func TestService_Watch_LazyInitsCursorToFirstBatchMinPosition(t *testing.T) {
	// Real-world positions may not start at 0 (e.g. 1-indexed, or numbered
	// per build session). The first non-empty batch's lowest position
	// becomes the cursor floor — otherwise the algorithm would stall forever
	// waiting for a position that will never arrive.
	var logCalls, buildCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-2/log":
			n := int(logCalls.Add(1))
			switch n {
			case 1:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"first\n","position":100},{"chunk":"second\n","position":101}],"next_after_timestamp":"ts1"}`))
			default:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[]}`))
			}
		case "/apps/my-app/builds/b-2":
			n := int(buildCalls.Add(1))
			if n == 1 {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-2","build_number":2,"status":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-2","build_number":2,"status":1,"triggered_workflow":"primary","branch":"main"}}`))
			}
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	_, err := svc.Watch(context.Background(), "my-app", "b-2", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "first\nsecond\n", buf.String(), "cursor must lazy-init to lowest position, not assume 0")
}

func TestService_Watch_StaleBufferDrainsAfterGapStalls(t *testing.T) {
	// If a gap never fills in (the missing chunk was dropped server-side or
	// never gets surfaced), the stale-buffer guard force-flushes after a few
	// polls so streaming keeps moving instead of stalling until the build
	// finishes.
	var logCalls, buildCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-3/log":
			n := int(logCalls.Add(1))
			switch n {
			case 1:
				// Position 3 + 5 with a gap (4 missing) — gap never fills.
				// Lazy init sets cursor to 3, so 3 emits immediately and 5
				// stays buffered waiting for 4 that never comes.
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"early\n","position":3},{"chunk":"late\n","position":5}],"next_after_timestamp":"ts1"}`))
			default:
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[],"next_after_timestamp":"ts1"}`))
			}
		case "/apps/my-app/builds/b-3":
			n := int(buildCalls.Add(1))
			if n <= 6 {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-3","build_number":3,"status":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"data":{"slug":"b-3","build_number":3,"status":1,"triggered_workflow":"primary","branch":"main"}}`))
			}
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	_, err := svc.Watch(context.Background(), "my-app", "b-3", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "early\nlate\n", buf.String(), "stale-buffer drain must emit both chunks in order")
}

func TestService_Watch_AlreadyArchived(t *testing.T) {
	rawSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ARCHIVED LOG\n"))
	}))
	t.Cleanup(rawSrv.Close)

	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-done/log":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"is_archived":          true,
				"expiring_raw_log_url": rawSrv.URL,
			})
		case "/apps/my-app/builds/b-done":
			_, _ = w.Write([]byte(`{"data":{"slug":"b-done","build_number":5,"status":1}}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	build, err := svc.Watch(context.Background(), "my-app", "b-done", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ARCHIVED LOG")
	assert.Equal(t, "success", build.Status)
}

func TestService_Watch_RetriesOn404(t *testing.T) {
	// First log call returns 404 (build not yet started); second succeeds.
	var callCount atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := int(callCount.Add(1))
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			if n == 1 {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"build not found"}`))
				return
			}
			// Second call: build has started, archived immediately.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"is_archived": true,
				"log_chunks":  []map[string]any{{"chunk": "log line\n", "position": 0}},
			})
		case "/apps/my-app/builds/b-1":
			_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":1}}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	build, err := svc.Watch(context.Background(), "my-app", "b-1", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "success", build.Status)
}

func TestService_Watch_FailsOnSecond404(t *testing.T) {
	// Both log calls return 404; Watch should return an error.
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/my-app/builds/b-1/log" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"build not found"}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	_, err := svc.Watch(context.Background(), "my-app", "b-1", io.Discard, time.Millisecond)
	require.Error(t, err)
	assert.EqualError(t, err, `build "b-1" not found`)
}

func TestService_Watch_StopsOnIsArchived(t *testing.T) {
	// Poll returns IsArchived=true with a non-empty NextAfterTimestamp;
	// Watch must stop and not loop indefinitely.
	var logCallCount atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			n := int(logCallCount.Add(1))
			if n == 1 {
				// First call: in-progress with a next timestamp.
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"line1\n","position":0}],"next_after_timestamp":"ts1"}`))
				return
			}
			// Second call: archived but NextAfterTimestamp still set — this is the bug scenario.
			_, _ = w.Write([]byte(`{"is_archived":true,"log_chunks":[{"chunk":"line2\n","position":1}],"next_after_timestamp":"ts2"}`))
		case "/apps/my-app/builds/b-1":
			_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":1,"status":1}}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	build, err := svc.Watch(context.Background(), "my-app", "b-1", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "line1")
	assert.Contains(t, buf.String(), "line2")
	assert.Equal(t, "success", build.Status)
}

func TestService_Watch_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		// Return a next timestamp so Watch will enter the sleep-poll loop,
		// then cancel the context so the select fires.
		_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"line\n","position":0}],"next_after_timestamp":"ts1"}`))
		cancel()
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	_, err := svc.Watch(ctx, "my-app", "b-1", &buf, 10*time.Millisecond)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestService_Watch_StopsOnBuildStatus(t *testing.T) {
	// Watch should exit the poll loop when the build status is no longer
	// in-progress (0), without waiting for the log to be archived.
	var logCalls, buildCalls atomic.Int32
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apps/my-app/builds/b-1/log":
			n := int(logCalls.Add(1))
			if n == 1 {
				// Initial call: in-progress, gives first chunks.
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"line1\n","position":0}],"next_after_timestamp":"ts1"}`))
			} else {
				// Poll: more chunks arrive but is_archived stays false.
				_, _ = w.Write([]byte(`{"is_archived":false,"log_chunks":[{"chunk":"line2\n","position":1}],"next_after_timestamp":"ts2"}`))
			}
		case "/apps/my-app/builds/b-1":
			buildCalls.Add(1)
			_, _ = w.Write([]byte(`{"data":{"slug":"b-1","build_number":5,"status":1,"status_text":"success"}}`))
		}
	})
	svc := NewService(newAPIClient(t, srv.URL))

	var buf bytes.Buffer
	build, err := svc.Watch(context.Background(), "my-app", "b-1", &buf, time.Millisecond)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "line1")
	assert.Contains(t, buf.String(), "line2")
	assert.Equal(t, "success", build.Status)
	assert.GreaterOrEqual(t, int(buildCalls.Load()), 1)
	// Should have stopped after a single poll cycle, not looped many times.
	assert.LessOrEqual(t, int(logCalls.Load()), 3, "Watch did not stop on build status")
}

func TestService_Watch_NilClientFails(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Watch(context.Background(), "a", "b", io.Discard, time.Second)
	assert.Error(t, err)
	_, err = svc.WaitForCompletion(context.Background(), "a", "b", time.Second)
	assert.Error(t, err)
}
