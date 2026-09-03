package claude

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil/picker"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
)

// errResumeCancelled signals the user backed out of the resume picker (empty
// input, "q", EOF, or Ctrl-C). The command treats it as a clean exit.
var errResumeCancelled = errors.New("resume cancelled")

// resumeOptions selects which previously-recorded session to resume.
type resumeOptions struct {
	target         string // explicit SESSION_ID/name ("" = none given)
	continueLatest bool
	waitTimeout    time.Duration
}

// runResume reconnects to a previous `rde claude` session recorded for the
// current repo. If the session is still running it reattaches; otherwise it
// restores the session and continues the same Claude Code conversation.
//
// A candidate is validated just-in-time — only the session the user actually
// chooses is looked up, so the picker stays instant and we never pull a whole
// workspace's session list to check a handful of IDs. If the chosen session is
// gone or can't be restored, we say so and drop the stale local record (never
// silently); for the picker / --continue we then fall through to the next
// candidate.
func runResume(ctx context.Context, cmd *cobra.Command, opts resumeOptions) error {
	log := newStepLogger(cmd)
	repoPath := repoRootPath(ctx)

	client, err := cmdutil.NewRDEClient(cmd)
	if err != nil {
		return err
	}
	svc := internalrde.NewService(client)

	for {
		rec, err := resolveResumeRecord(ctx, cmd, svc, repoPath, opts)
		if errors.Is(err, errResumeCancelled) {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
			return nil
		}
		if err != nil {
			return err
		}

		log.group("Session")
		log.step("Resuming %q (%s)…", rec.DisplayName(), rec.RDESessionID)

		sess, err := svc.GetSession(ctx, rec.WorkspaceID, rec.RDESessionID)
		if err != nil {
			if internalrde.IsNotFound(err) {
				retry, herr := handleUnresumable(log, repoPath, rec, "no longer exists", opts)
				if retry {
					continue
				}
				return herr
			}
			return fmt.Errorf("look up session %s: %w", rec.RDESessionID, err)
		}
		if !sess.Resumable() {
			retry, herr := handleUnresumable(log, repoPath, rec, "can't be restored (its persistent disk is no longer available)", opts)
			if retry {
				continue
			}
			return herr
		}

		return resumeSession(ctx, cmd, svc, log, rec, sess, opts)
	}
}

// handleUnresumable drops the stale local record for a session that's gone or
// unrestorable — always telling the user, never silently. It reports whether
// the caller should try another candidate: the picker and --continue fall
// through to the next one; an explicit SESSION_ID stops with an error.
func handleUnresumable(log *stepLogger, repoPath string, rec localsession.Record, reason string, opts resumeOptions) (retry bool, err error) {
	if rmErr := localsession.Remove(repoPath, rec.RDESessionID); rmErr != nil {
		log.warn("Could not remove the local record for %s: %v", rec.RDESessionID, rmErr)
	}
	if opts.target == "" {
		log.warn("Session %q (%s) %s — removed it from your resume list.", rec.DisplayName(), rec.RDESessionID, reason)
		return true, nil
	}
	return false, fmt.Errorf("session %s %s; removed it from your resume list", rec.RDESessionID, reason)
}

// resumeSession reconnects to (or restores) an existing, resumable session and
// attaches to its Claude Code conversation.
func resumeSession(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, log *stepLogger, rec localsession.Record, sess internalrde.Session, opts resumeOptions) error {
	waitCtx, cancel := context.WithTimeout(ctx, opts.waitTimeout)
	defer cancel()

	switch sess.Status {
	case "running":
		// Still alive (e.g. after a disconnect or laptop sleep) — reattach
		// directly, no restore needed.
	case "terminated", "stopped", "failed":
		log.step("Restoring session…")
		if _, err := svc.RestoreSession(ctx, rec.WorkspaceID, rec.RDESessionID); err != nil {
			return fmt.Errorf("restore session: %w", err)
		}
		var ready internalrde.Session
		if err := log.await(waitCtx, "Booting session…", "Session booted",
			func(c context.Context, status func(string)) error {
				var e error
				ready, e = svc.WaitForReady(c, rec.WorkspaceID, rec.RDESessionID, 0, status)
				return e
			}); err != nil {
			return fmt.Errorf("waiting for session: %w", err)
		}
		if ready.Status != "running" {
			cmdutil.SilenceRootErrors(cmd)
			return fmt.Errorf("session ended provisioning with status %q (expected running)", ready.Status)
		}
	default:
		cmdutil.SilenceRootErrors(cmd)
		return fmt.Errorf("session %s is %q and can't be resumed right now; try again shortly", rec.RDESessionID, sess.Status)
	}

	// Resume keeps the terminate-on-exit lifecycle, same as a fresh run, so
	// VMs don't linger after you detach. The exception is leaveRunning: when the
	// user interrupts a reconnect, the VM is left running so they can reconnect
	// later.
	leaveRunning := false
	defer terminateOnExit(svc, log, rec.WorkspaceID, rec.RDESessionID, rec.DisplayName(), &leaveRunning)()

	if err := log.await(waitCtx, "Waiting for remote access…", "Remote access ready",
		func(c context.Context, _ func(string)) error {
			_, e := svc.WaitForSSHReady(c, rec.WorkspaceID, rec.RDESessionID, 0)
			return e
		}); err != nil {
		return fmt.Errorf("waiting for SSH access: %w", err)
	}
	log.done("Session ready")

	// Re-establish SSH agent forwarding so git-over-SSH inside the resumed
	// session keeps working. The repo is already on the persistent disk, so
	// there's no clone step here.
	cleanupAgent, repoAuth, _ := ensureAgentHasKey(ctx, log, gitSSHURL(rec.Repo))
	defer cleanupAgent()
	log.step("Auth: %s", repoAuth)

	// Re-apply the user's local git identity (idempotent, global): the repo's
	// on the persistent disk, but their identity may have changed locally, and
	// sessions created before this existed never had it set.
	syncGitIdentity(ctx, svc, log, rec.WorkspaceID, rec.RDESessionID)

	// ── Claude Code ────────────────────────────────────────────────
	log.group("Claude Code")
	log.step("Resuming…")
	exitCode, err := attachClaude(ctx, svc, log, attachParams{
		workspaceID:     rec.WorkspaceID,
		sessionID:       rec.RDESessionID,
		claudeSessionID: rec.ClaudeSessionID,
		claudeCmd:       buildResumeCommand(rec.RemoteRepoDir, rec.ClaudeSessionID),
		record:          rec,
		describe:        newDescriber(repoSlugFromURL(rec.Repo), rec.Branch),
	})
	if errors.Is(err, errReconnectInterrupted) {
		leaveRunning = true // skip termination; deferred cleanup leaves the VM up
		return nil
	}
	if err != nil {
		return err
	}
	return claudeExitError(cmd, exitCode)
}

// resolveResumeRecord picks the record to resume from the options: the latest
// for --continue, an exact match for an explicit target, or an interactive
// picker otherwise.
func resolveResumeRecord(ctx context.Context, cmd *cobra.Command, getter sessionStatusGetter, repoPath string, opts resumeOptions) (localsession.Record, error) {
	if opts.continueLatest && opts.target != "" {
		return localsession.Record{}, fmt.Errorf("cannot combine --continue with a SESSION_ID")
	}
	switch {
	case opts.continueLatest:
		rec, ok, err := localsession.Latest(repoPath)
		if err != nil {
			return localsession.Record{}, fmt.Errorf("read local sessions: %w", err)
		}
		if !ok {
			return localsession.Record{}, errNoSessions
		}
		return rec, nil
	case opts.target != "":
		return findRecord(repoPath, opts.target)
	default:
		return pickRecord(ctx, cmd, getter, repoPath)
	}
}

var errNoSessions = fmt.Errorf("no previous rde claude session found for this repo; run 'rde claude' to start one")

// sessionStatusGetter is the slice of the rde service the picker needs to show
// each candidate's live status. Narrowed to an interface so the status logic is
// testable without a real client.
type sessionStatusGetter interface {
	GetSession(ctx context.Context, workspaceID, sessionID string) (internalrde.Session, error)
}

// statusFetchTimeout caps how long the picker waits for live statuses. Past it,
// the still-pending entries just show "status unknown" — they stay selectable,
// and the chosen one is validated for real before we act on it.
const statusFetchTimeout = 8 * time.Second

// recordStatus is a record's live status as shown in the picker.
type recordStatus struct {
	status    string // live status word, "deleted", or "" when the check failed/timed out
	resumable bool
}

// statusOf maps a GetSession result to a picker status. A 404 is "deleted"; any
// other error leaves the status unknown (assumed resumable — the just-in-time
// check at selection time decides for real).
func statusOf(sess internalrde.Session, err error) recordStatus {
	switch {
	case err == nil:
		return recordStatus{status: sess.Status, resumable: sess.Resumable()}
	case internalrde.IsNotFound(err):
		return recordStatus{status: "deleted", resumable: false}
	default:
		return recordStatus{resumable: true}
	}
}

// statusLabel renders a recordStatus for the picker list.
func statusLabel(rs recordStatus) string {
	switch {
	case rs.status == "":
		return "status unknown"
	case rs.status == "deleted":
		return "deleted"
	case !rs.resumable:
		return rs.status + " · unrestorable"
	default:
		return rs.status
	}
}

// statusTone maps a recordStatus to a picker color tone, mirroring the session
// status palette: running→green, deleted/failed→red, unrestorable→amber, and
// everything else (stopped/terminated/unknown)→grey.
func statusTone(rs recordStatus) picker.Tone {
	switch {
	case rs.status == "":
		return picker.ToneDim // status unknown
	case rs.status == "deleted":
		return picker.ToneDanger
	case !rs.resumable:
		return picker.ToneWarn // "… · unrestorable"
	case rs.status == "running":
		return picker.ToneSuccess
	case rs.status == "failed":
		return picker.ToneDanger
	default:
		return picker.ToneDim
	}
}

// fetchStatuses looks up every record's live status concurrently (bounded), so
// the picker shows status without paying N sequential round-trips. Results are
// index-aligned with recs; a slow API degrades to "status unknown" rather than
// stalling the picker.
func fetchStatuses(ctx context.Context, getter sessionStatusGetter, recs []localsession.Record) []recordStatus {
	ctx, cancel := context.WithTimeout(ctx, statusFetchTimeout)
	defer cancel()

	out := make([]recordStatus, len(recs))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, r := range recs {
		wg.Add(1)
		go func(i int, r localsession.Record) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sess, err := getter.GetSession(ctx, r.WorkspaceID, r.RDESessionID)
			out[i] = statusOf(sess, err)
		}(i, r)
	}
	wg.Wait()
	return out
}

// findRecord resolves an explicit target to a record, matching the RDE session
// ID first, then the session name or AI title (case-insensitive).
func findRecord(repoPath, target string) (localsession.Record, error) {
	recs, err := localsession.ListByProject(repoPath)
	if err != nil {
		return localsession.Record{}, fmt.Errorf("read local sessions: %w", err)
	}
	for _, r := range recs {
		if r.RDESessionID == target {
			return r, nil
		}
	}
	var matches []localsession.Record
	for _, r := range recs {
		if strings.EqualFold(r.Name, target) || strings.EqualFold(r.AITitle, target) {
			matches = append(matches, r)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return localsession.Record{}, fmt.Errorf("no session %q found for this repo (run 'rde claude --resume' to see the list)", target)
	default:
		return localsession.Record{}, fmt.Errorf("%q matches %d sessions for this repo; resume by ID instead", target, len(matches))
	}
}

// pickRecord shows the repo's recorded sessions in an interactive picker, with
// each row's live status fetched in parallel for display. The cursor starts on
// the most recent (records are newest-first), so Enter resumes it; Esc, "q", or
// Ctrl-C back out with errResumeCancelled. It errors (rather than hanging) when
// stdin/stderr isn't a terminal.
func pickRecord(ctx context.Context, cmd *cobra.Command, getter sessionStatusGetter, repoPath string) (localsession.Record, error) {
	recs, err := localsession.ListByProject(repoPath)
	if err != nil {
		return localsession.Record{}, fmt.Errorf("read local sessions: %w", err)
	}
	if len(recs) == 0 {
		return localsession.Record{}, errNoSessions
	}
	if !cmdutil.IsTerminal(os.Stdin) || !cmdutil.IsTerminalWriter(cmd.ErrOrStderr()) {
		return localsession.Record{}, fmt.Errorf("--resume needs a SESSION_ID when input is not a terminal; pass 'rde claude --resume <id>'")
	}

	// Look up live statuses in parallel for display. Deleted/unrestorable
	// entries are shown marked (not removed — that would be jarring); the stale
	// record is cleared only if the user actually selects it, where the
	// just-in-time check confirms it for real.
	statuses := fetchStatuses(ctx, getter, recs)

	items := make([]picker.Item, len(recs))
	for i, r := range recs {
		items[i] = picker.Item{
			Title:  describeRecord(r),
			Status: statusLabel(statuses[i]),
			Tone:   statusTone(statuses[i]),
		}
	}
	// The cursor starts on the newest session (row 0); the resume list can grow
	// long, so enable type-to-filter on the title/branch.
	idx, err := picker.Select(ctx, picker.Config{
		Prompt:     "Select a session to resume",
		Items:      items,
		Cursor:     0,
		DefaultIdx: 0,
		Filter:     true,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		return localsession.Record{}, errResumeCancelled
	}
	if err != nil {
		return localsession.Record{}, err
	}
	return recs[idx], nil
}

// describeRecord is a one-line summary for the picker: title, branch, and how
// long ago it was last active.
func describeRecord(r localsession.Record) string {
	parts := []string{r.DisplayName()}
	if r.Branch != "" {
		parts = append(parts, "["+r.Branch+"]")
	}
	parts = append(parts, ago(r.UpdatedAt))
	return strings.Join(parts, "  ")
}

// ago renders a coarse "time since" for the picker.
func ago(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
