package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// Watch streams the build log to w using the API's delta-log protocol,
// blocking until the build finishes. Returns the final Build so the caller
// can determine the exit code. Ctrl-C (context cancellation) causes Watch to
// return context.Canceled; the build continues running on Bitrise.
//
// For builds that are already finished when Watch is called, the archived
// log is fetched and streamed immediately.
func (s *Service) Watch(ctx context.Context, appSlug, buildSlug string, w io.Writer, interval time.Duration) (Build, error) {
	if s.client == nil {
		return Build{}, fmt.Errorf("API client not configured")
	}
	if appSlug == "" {
		return Build{}, fmt.Errorf("app ID is required")
	}
	if buildSlug == "" {
		return Build{}, fmt.Errorf("build ID is required")
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}

	// The log manifest returns 404 for up to ~15s after trigger while the
	// runner provisions. Retry on 404 for up to 10 intervals (~30s at the
	// default 3s interval) before giving up.
	const maxInitialRetries = 10
	var (
		manifest bitriseapi.BuildLogResponse
		err      error
	)
	for attempt := 0; ; attempt++ {
		manifest, err = s.client.BuildLogManifest(ctx, appSlug, buildSlug, "")
		if err == nil {
			break
		}
		var apiErr *bitriseapi.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			return Build{}, err
		}
		if attempt >= maxInitialRetries {
			return Build{}, fmt.Errorf("build %q not found", buildSlug)
		}
		select {
		case <-ctx.Done():
			return Build{}, ctx.Err()
		case <-time.After(interval):
		}
	}

	// Already finished: stream the archived log via the existing Log path.
	if manifest.IsArchived || manifest.ExpiringRawLogURL != "" {
		if err := s.Log(ctx, appSlug, buildSlug, w); err != nil {
			return Build{}, err
		}
		return s.View(ctx, appSlug, buildSlug)
	}

	// In-progress: maintain a position-keyed buffer across polls so chunks
	// delivered out of order can be reordered before reaching the writer.
	// The Bitrise log API returns chunks in arbitrary order across polls
	// because the server collects them from parallel producers; without
	// cross-poll buffering, a chunk at position N can land on screen before
	// position N-1 arrives in the next poll.
	//
	// The cursor (nextEmit) is lazy-initialized to the lowest position in
	// the first non-empty batch — we don't assume positions start at zero,
	// since the API doesn't document the numbering and may be 1-indexed.
	// A "stale buffer" guard force-drains after maxStaleFlushes polls with
	// no progress: positions can have real gaps (parallel sections, dropped
	// chunks), and waiting for a chunk that never arrives would stall
	// streaming until the build finishes.
	const maxStaleFlushes = 3
	buffer := map[int]string{}
	nextEmit := -1
	staleFlushes := 0
	flush := func(batch []bitriseapi.BuildLogChunk) error {
		emitted, err := flushContiguous(w, buffer, &nextEmit, batch)
		if err != nil {
			return err
		}
		if emitted > 0 {
			staleFlushes = 0
			return nil
		}
		if len(buffer) == 0 {
			return nil
		}
		staleFlushes++
		if staleFlushes < maxStaleFlushes {
			return nil
		}
		if err := drainRemaining(w, buffer, &nextEmit); err != nil {
			return err
		}
		staleFlushes = 0
		return nil
	}

	if err := flush(manifest.LogChunks); err != nil {
		return Build{}, err
	}

	lastAfterTimestamp := ""
	afterTimestamp := manifest.NextAfterTimestamp
	var current bitriseapi.Build
	for {
		select {
		case <-ctx.Done():
			return Build{}, ctx.Err()
		case <-time.After(interval):
		}
		current, err = s.client.Build(ctx, appSlug, buildSlug)
		if err != nil {
			return Build{}, err
		}
		if afterTimestamp != "" {
			manifest, err = s.client.BuildLogManifest(ctx, appSlug, buildSlug, afterTimestamp)
			if err != nil {
				return Build{}, err
			}
			if err := flush(manifest.LogChunks); err != nil {
				return Build{}, err
			}
			lastAfterTimestamp = afterTimestamp
			afterTimestamp = manifest.NextAfterTimestamp
		}
		if current.Status != 0 {
			break
		}
	}

	// One final call to flush any chunks buffered after the last poll. Runs
	// unconditionally, even if lastAfterTimestamp is still "" (a poll's
	// NextAfterTimestamp can come back empty), so a stuck-empty cursor can't
	// permanently skip this catch-up — worst case it's a full log refetch,
	// and flushContiguous drops anything already emitted.
	final, err := s.client.BuildLogManifest(ctx, appSlug, buildSlug, lastAfterTimestamp)
	if err != nil {
		return Build{}, err
	}
	if _, err := flushContiguous(w, buffer, &nextEmit, final.LogChunks); err != nil {
		return Build{}, err
	}

	// Failsafe drain: any chunks still buffered (a position before them
	// never arrived from the server) get emitted in position order. Better
	// to surface them with a gap than to silently drop log lines.
	if err := drainRemaining(w, buffer, &nextEmit); err != nil {
		return Build{}, err
	}

	return fromAPI(current, appSlug), nil
}

// flushContiguous merges batch into buffer and writes any chunks whose
// position extends the contiguous run from *nextEmit. Returns the number of
// chunks emitted. Out-of-order chunks stay in buffer until the gap before
// them fills in via a later poll or the stale-buffer drain.
//
// *nextEmit == -1 signals "not yet initialized"; the first non-empty batch
// initializes it to the lowest position seen so the algorithm works
// regardless of whether the API uses 0-indexed or 1-indexed positions.
//
// Replay handling:
//   - position < *nextEmit (after init): already emitted, drop silently
//   - position already in buffer: keep the first-arrived copy
func flushContiguous(w io.Writer, buffer map[int]string, nextEmit *int, batch []bitriseapi.BuildLogChunk) (int, error) {
	for _, c := range batch {
		if *nextEmit >= 0 && c.Position < *nextEmit {
			continue
		}
		if _, exists := buffer[c.Position]; exists {
			continue
		}
		buffer[c.Position] = c.Chunk
	}
	if *nextEmit < 0 {
		if len(buffer) == 0 {
			return 0, nil
		}
		min := -1
		for p := range buffer {
			if min < 0 || p < min {
				min = p
			}
		}
		*nextEmit = min
	}
	emitted := 0
	for {
		chunk, ok := buffer[*nextEmit]
		if !ok {
			return emitted, nil
		}
		if _, err := io.WriteString(w, chunk); err != nil {
			return emitted, fmt.Errorf("write log chunk: %w", err)
		}
		delete(buffer, *nextEmit)
		*nextEmit++
		emitted++
	}
}

// drainRemaining writes any chunks still in buffer in position order and
// advances *nextEmit past the highest drained position so future chunks at
// positions below it are dropped as already-emitted. Called both as the
// final failsafe at build end and as the stale-buffer escape hatch when a
// gap has stalled streaming for too many polls.
func drainRemaining(w io.Writer, buffer map[int]string, nextEmit *int) error {
	if len(buffer) == 0 {
		return nil
	}
	positions := make([]int, 0, len(buffer))
	for p := range buffer {
		positions = append(positions, p)
	}
	slices.Sort(positions)
	for _, p := range positions {
		if _, err := io.WriteString(w, buffer[p]); err != nil {
			return fmt.Errorf("write log chunk: %w", err)
		}
		delete(buffer, p)
	}
	*nextEmit = positions[len(positions)-1] + 1
	return nil
}

// WaitForCompletion blocks until the build is no longer in-progress, polling
// at the given interval. Returns the final Build. Ctrl-C (context
// cancellation) returns context.Canceled; the build keeps running on
// Bitrise.
func (s *Service) WaitForCompletion(ctx context.Context, appSlug, buildSlug string, interval time.Duration) (Build, error) {
	if s.client == nil {
		return Build{}, fmt.Errorf("API client not configured")
	}
	if appSlug == "" {
		return Build{}, fmt.Errorf("app ID is required")
	}
	if buildSlug == "" {
		return Build{}, fmt.Errorf("build ID is required")
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for {
		b, err := s.View(ctx, appSlug, buildSlug)
		if err != nil {
			return Build{}, err
		}
		if b.Status != "in-progress" {
			return b, nil
		}
		select {
		case <-ctx.Done():
			return Build{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}
