package rde

import (
	"context"
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// Friendly --stage values accepted by the CLI. The backend takes the numeric
// LogStage enum; these map to it in logStageToAPI. The backend calls the
// second stage "main" internally; the user-facing name is "startup" (it runs
// the session's startup script).
const (
	LogStageWarmup  = "warmup"  // runs once at session creation
	LogStageStartup = "startup" // runs on every session start/restart
)

// LogStages is the ordered set of valid --stage values, for validation and
// shell completion.
var LogStages = []string{LogStageWarmup, LogStageStartup}

// logStageToAPI maps a friendly stage name to the numeric LogStage enum string
// the backend expects in the path (1=warmup, 2=main/startup).
func logStageToAPI(stage string) (string, error) {
	switch stage {
	case LogStageWarmup:
		return "1", nil
	case LogStageStartup:
		return "2", nil
	default:
		return "", fmt.Errorf("invalid stage %q (must be %s or %s)", stage, LogStageWarmup, LogStageStartup)
	}
}

// StreamSessionLogs streams one stage's log for a session, invoking fn for each
// content chunk's text in order. stage is a friendly name (warmup/startup).
//
// idleTimeout controls when to stop: 0 follows live until ctx is cancelled
// (Ctrl-C); a positive value returns once no new content has arrived for that
// long, which delivers the replayed log-so-far and then exits.
//
// A pre-stream "logs not ready" condition surfaces as a 404 *rdeapi.APIError,
// which the cli command layer distinguishes to decide between a friendly exit
// and a follow-mode retry.
func (s *Service) StreamSessionLogs(ctx context.Context, workspaceID, sessionID, stage string, idleTimeout time.Duration, fn func(string) error) error {
	if s.client == nil {
		return errClient()
	}
	apiStage, err := logStageToAPI(stage)
	if err != nil {
		return err
	}
	return s.client.StreamSessionLogs(ctx, workspaceID, sessionID, apiStage, idleTimeout, func(chunk rdeapi.LogChunk) error {
		return fn(chunk.LogContent)
	})
}
