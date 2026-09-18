package rde

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// ErrConnectionLost marks an interactive run that ended because the connection
// to the session dropped (network failure, dropped SSH channel) rather than
// the remote program exiting. Callers can match it with errors.Is to decide
// whether to reconnect. A program inside a survivable wrapper (e.g. tmux) keeps
// running on the session, so reattaching resumes it.
var ErrConnectionLost = errors.New("connection to the session was lost")

// ExecuteInteractive attaches the caller's terminal to command running on the
// session over SSH and blocks until it exits, returning its exit code. Unlike
// Execute, it allocates a PTY (when stdin is a terminal), runs in raw mode, and
// is NOT capped by Execute's timeout — interactive programs are long-lived.
//
// It is the interactive sibling of Execute: same pre-flight checks, same SSH
// dial and agent-forwarding posture, but stdin/stdout/stderr are streamed live
// instead of captured.
func (s *Service) ExecuteInteractive(ctx context.Context, workspaceID, sessionID, command string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	if s.client == nil {
		return -1, errClient()
	}
	if command == "" {
		return -1, fmt.Errorf("command is required")
	}
	sess, err := s.GetSession(ctx, workspaceID, sessionID)
	if err != nil {
		// A network-level failure reaching the API (e.g. during an outage) is
		// recoverable — surface it as a lost connection so callers can retry.
		if isRetryableDialErr(err) {
			return -1, fmt.Errorf("%w: %v", ErrConnectionLost, err)
		}
		return -1, fmt.Errorf("fetch session: %w", err)
	}
	target, err := sshTargetForSession(sess)
	if err != nil {
		return -1, err
	}

	// Retry the dial: the backend reports SSH ready a moment before the port
	// actually accepts connections, so the first attempts can be refused.
	client, err := dialSSHWithRetry(ctx, target)
	if err != nil {
		// A transient connection failure (the dial-retry budget elapsed while
		// the port/network was down) is recoverable; an auth/handshake failure
		// is not.
		if isRetryableDialErr(err) {
			return -1, fmt.Errorf("%w: %v", ErrConnectionLost, err)
		}
		return -1, err
	}
	defer client.Close() //nolint:errcheck // run errors take precedence; nothing actionable on close failure

	return client.runInteractive(ctx, command, stdin, stdout, stderr)
}
