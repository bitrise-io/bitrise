package rdeapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LogChunk is one frame of a session log stream. HeartbeatMessage frames
// carry empty LogContent and exist only to keep the connection alive; callers
// skip them.
type LogChunk struct {
	LogContent       string `json:"logContent"`
	HeartbeatMessage bool   `json:"heartbeatMessage"`
}

// maxLogLine caps the bufio.Scanner token size. Log lines can be long (stack
// traces, base64 blobs); the default 64KiB is too small, so allow up to 1MiB.
const maxLogLine = 1 << 20

// StreamSessionLogs opens the server-streaming log endpoint for one stage of a
// session and invokes fn for every non-empty content chunk, in order.
//
// Endpoint: GET /v1/workspaces/{workspaceId}/sessions/{sessionId}/logs/{stage}
// where stage is the numeric LogStage enum ("1"=warmup, "2"=main). The stream
// replays the stage log from the start, then continues live. The backend never
// sends EOF when the stage's script finishes — it holds the connection open
// with minute-interval heartbeats — so callers control how long to listen:
//
//   - idleTimeout <= 0: stream until the connection actually closes or ctx is
//     cancelled (Ctrl-C). Use for a live "follow".
//   - idleTimeout  > 0: return cleanly once no new content has arrived for that
//     long. The replayed backlog arrives in a burst, so this delivers the whole
//     log-so-far and then exits ("print what's there, don't wait").
//
// Heartbeat/empty frames are skipped and do not count as content. A mid-stream
// {"error":...} frame is surfaced as an error. A cancelled ctx is treated as a
// clean end of stream and returns nil. Pre-stream failures come back as
// *APIError (e.g. 404 = "logs not available yet").
func (c *Client) StreamSessionLogs(ctx context.Context, workspaceID, sessionID, stage string, idleTimeout time.Duration, fn func(LogChunk) error) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/logs/"+url.PathEscape(stage))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+p, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.doStream(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// frame is one parsed result off the wire: a content chunk, a terminal
	// error, or done (the stream closed without error).
	type frame struct {
		chunk LogChunk
		err   error
		done  bool
	}
	frames := make(chan frame)
	stop := make(chan struct{})
	defer close(stop) // unblocks the reader goroutine when we return early

	go func() {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), maxLogLine)
		send := func(f frame) bool {
			select {
			case frames <- f:
				return true
			case <-stop:
				return false
			}
		}
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(strings.TrimSpace(string(line))) == 0 {
				continue
			}
			var ln logStreamLine
			if jerr := json.Unmarshal(line, &ln); jerr != nil {
				send(frame{err: fmt.Errorf("decode log frame: %w", jerr)})
				return
			}
			if ln.Error != nil {
				send(frame{err: streamFrameError(ln.Error)})
				return
			}
			if ln.Result.HeartbeatMessage || ln.Result.LogContent == "" {
				continue // heartbeats/empty frames are not content
			}
			if !send(frame{chunk: ln.Result}) {
				return
			}
		}
		if serr := scanner.Err(); serr != nil {
			send(frame{err: serr})
			return
		}
		send(frame{done: true})
	}()

	// Idle timer: armed from the start and reset on every content chunk, so it
	// fires once the replayed backlog has drained. Left nil (never fires) when
	// idleTimeout <= 0.
	var idleTimer *time.Timer
	var idleC <-chan time.Time
	if idleTimeout > 0 {
		idleTimer = time.NewTimer(idleTimeout)
		defer idleTimer.Stop()
		idleC = idleTimer.C
	}
	resetIdle := func() {
		if idleTimer == nil {
			return
		}
		if !idleTimer.Stop() {
			select {
			case <-idleTimer.C:
			default:
			}
		}
		idleTimer.Reset(idleTimeout)
	}

	for {
		select {
		case <-ctx.Done():
			return nil // Ctrl-C / cancellation is a clean stop
		case <-idleC:
			return nil // no new content within idleTimeout — backlog drained
		case f := <-frames:
			switch {
			case f.err != nil:
				if ctx.Err() != nil {
					return nil
				}
				return f.err
			case f.done:
				return nil
			default:
				if err := fn(f.chunk); err != nil {
					return err
				}
				resetIdle()
			}
		}
	}
}

// logStreamLine is one newline-delimited JSON frame emitted by the
// grpc-gateway ForwardResponseStream wrapper. A frame carries EITHER a result
// or an error: success frames look like {"result":{...}} and a mid-stream
// failure looks like {"error":{"code":int,"message":string,"details":[...]}}
// (a google.rpc.Status — code is a gRPC code, not HTTP).
type logStreamLine struct {
	Result LogChunk   `json:"result"`
	Error  *errorBody `json:"error"`
}

// streamFrameError renders a mid-stream {"error":...} frame as a plain error.
// Unlike a pre-stream HTTP failure, code here is a gRPC code (not an HTTP
// status), so this deliberately omits the "RDE API <status>" prefix and just
// surfaces the message plus any field violations.
func streamFrameError(e *errorBody) error {
	msg := e.Message
	if v := strings.Join(violationsFromDetails(e.Details), "; "); v != "" {
		if msg != "" {
			msg += ": " + v
		} else {
			msg = v
		}
	}
	if msg == "" {
		msg = "log stream error"
	}
	return errors.New(msg)
}
