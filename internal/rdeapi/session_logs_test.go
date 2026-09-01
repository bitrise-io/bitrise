package rdeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamSessionLogs_DeliversChunksAndSkipsHeartbeats(t *testing.T) {
	frames := []string{
		`{"result":{"logContent":"line one\n"}}`,
		`{"result":{"heartbeatMessage":true}}`,                  // skipped
		`{"result":{"logContent":"","heartbeatMessage":false}}`, // empty -> skipped
		`{"result":{"logContent":"line two\n"}}`,
	}
	c, gotPath := streamingServer(t, frames, http.StatusOK, nil)

	var got []string
	err := c.StreamSessionLogs(context.Background(), "ws-1", "s1", "2", 0, func(chunk LogChunk) error {
		got = append(got, chunk.LogContent)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamSessionLogs: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/logs/2"; *gotPath != want {
		t.Errorf("path = %q, want %q", *gotPath, want)
	}
	if len(got) != 2 || got[0] != "line one\n" || got[1] != "line two\n" {
		t.Errorf("chunks = %q, want [line one, line two]", got)
	}
}

func TestStreamSessionLogs_NotReady404(t *testing.T) {
	c, _ := streamingServer(t, []string{`{"code":5,"message":"logs not available yet for this stage"}`}, http.StatusNotFound, nil)

	err := c.StreamSessionLogs(context.Background(), "ws-1", "s1", "1", 0, func(LogChunk) error { return nil })
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestStreamSessionLogs_MidStreamErrorFrame(t *testing.T) {
	frames := []string{
		`{"result":{"logContent":"some output\n"}}`,
		`{"error":{"code":14,"message":"log stream interrupted, please reconnect"}}`,
	}
	c, _ := streamingServer(t, frames, http.StatusOK, nil)

	var got []string
	err := c.StreamSessionLogs(context.Background(), "ws-1", "s1", "2", 0, func(chunk LogChunk) error {
		got = append(got, chunk.LogContent)
		return nil
	})
	if err == nil {
		t.Fatal("expected error from mid-stream error frame")
	}
	if !strings.Contains(err.Error(), "please reconnect") {
		t.Errorf("error = %q, want it to mention 'please reconnect'", err.Error())
	}
	// The content delivered before the error frame is still surfaced.
	if len(got) != 1 || got[0] != "some output\n" {
		t.Errorf("chunks before error = %q", got)
	}
	// A mid-stream gRPC error is NOT an HTTP APIError.
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Errorf("mid-stream error should not be *APIError, got %v", apiErr)
	}
}

func TestStreamSessionLogs_CancelledContextIsCleanEOF(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	c, _ := streamingServer(t, []string{`{"result":{"logContent":"first\n"}}`}, http.StatusOK, block)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []string
	err := c.StreamSessionLogs(ctx, "ws-1", "s1", "2", 0, func(chunk LogChunk) error {
		got = append(got, chunk.LogContent)
		cancel() // simulate Ctrl-C right after the first chunk
		return nil
	})
	if err != nil {
		t.Fatalf("cancelled context should yield clean EOF (nil), got %v", err)
	}
	if len(got) != 1 || got[0] != "first\n" {
		t.Errorf("chunks = %q, want [first]", got)
	}
}

func TestStreamSessionLogs_IdleTimeoutStops(t *testing.T) {
	// Server writes the backlog then holds the connection open with no EOF,
	// like the real backend. With a positive idleTimeout the call must deliver
	// the backlog and then return cleanly once the stream goes quiet.
	block := make(chan struct{})
	defer close(block)
	c, _ := streamingServer(t, []string{
		`{"result":{"logContent":"a"}}`,
		`{"result":{"logContent":"b"}}`,
	}, http.StatusOK, block)

	var got []string
	err := c.StreamSessionLogs(context.Background(), "ws-1", "s1", "2", 30*time.Millisecond, func(chunk LogChunk) error {
		got = append(got, chunk.LogContent)
		return nil
	})
	if err != nil {
		t.Fatalf("idle stop should return nil, got %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("chunks = %q, want [a b] before idle stop", got)
	}
}

func TestStreamSessionLogs_ValidationGuards(t *testing.T) {
	c, err := New("https://unused", "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if err := c.StreamSessionLogs(ctx, "", "s1", LogStageMain, 0, func(LogChunk) error { return nil }); err == nil {
		t.Error("expected error for empty workspace ID")
	}
	if err := c.StreamSessionLogs(ctx, "ws", "", LogStageMain, 0, func(LogChunk) error { return nil }); err == nil {
		t.Error("expected error for empty session ID")
	}
	if err := c.StreamSessionLogs(ctx, "ws", "s1", LogStage("3"), 0, func(LogChunk) error { return nil }); err == nil {
		t.Error("expected error for invalid log stage")
	}
}

func TestStreamSessionLogs_DeadlineExceededIsAnError(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	c, _ := streamingServer(t, []string{`{"result":{"logContent":"first\n"}}`}, http.StatusOK, block)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := c.StreamSessionLogs(ctx, "ws-1", "s1", LogStageMain, 0, func(LogChunk) error { return nil })
	if err == nil {
		t.Fatal("expected an error when the context deadline is exceeded mid-stream, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want it to wrap context.DeadlineExceeded", err)
	}
}

// streamingServer serves the given newline-delimited frames on the logs
// endpoint, flushing each so the client sees them incrementally. It records
// the request path. If block is non-nil the handler waits on it (then on the
// request context) after the frames, simulating a still-open live stream.
func streamingServer(t *testing.T, frames []string, status int, block <-chan struct{}) (*Client, *string) {
	t.Helper()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if status != http.StatusOK {
			w.WriteHeader(status)
		}
		for _, f := range frames {
			_, _ = w.Write([]byte(f + "\n"))
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
		}
		if block != nil {
			select {
			case <-block:
			case <-r.Context().Done():
			}
		}
	}))
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, "tok")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, &gotPath
}
