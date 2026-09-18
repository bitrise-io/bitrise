package claude

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil/picker"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

func TestAgo(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		t    time.Time
		want string
	}{
		{time.Time{}, "unknown"},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-3 * time.Hour), "3h ago"},
		{now.Add(-50 * time.Hour), "2d ago"},
	} {
		if got := ago(tc.t); got != tc.want {
			t.Errorf("ago(%v) = %q, want %q", tc.t, got, tc.want)
		}
	}
}

func TestDescribeRecord(t *testing.T) {
	r := localsession.Record{Name: "claude-x", AITitle: "Fix bug", Branch: "main", UpdatedAt: time.Now().Add(-2 * time.Minute)}
	got := describeRecord(r)
	if got != "Fix bug  [main]  2m ago" {
		t.Errorf("describeRecord = %q", got)
	}
	// No branch → branch segment omitted.
	r2 := localsession.Record{Name: "claude-y", UpdatedAt: time.Now()}
	if got := describeRecord(r2); got != "claude-y  just now" {
		t.Errorf("describeRecord (no branch) = %q", got)
	}
}

func TestResolveResumeRecord(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"
	mustSave(t, localsession.Record{RDESessionID: "a", WorkspaceID: "ws", Name: "claude-a", RepoPath: repoPath})
	time.Sleep(5 * time.Millisecond)
	mustSave(t, localsession.Record{RDESessionID: "b", WorkspaceID: "ws", Name: "claude-b", AITitle: "Bee", RepoPath: repoPath})

	cmd := &cobra.Command{}

	// --continue → most recent (b).
	rec, err := resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{continueLatest: true}, nil)
	if err != nil || rec.RDESessionID != "b" {
		t.Errorf("continue = %q err=%v, want b", rec.RDESessionID, err)
	}

	// target by exact ID.
	rec, err = resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{target: "a"}, nil)
	if err != nil || rec.RDESessionID != "a" {
		t.Errorf("target id = %q err=%v, want a", rec.RDESessionID, err)
	}

	// target by AI title (case-insensitive).
	rec, err = resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{target: "bee"}, nil)
	if err != nil || rec.RDESessionID != "b" {
		t.Errorf("target title = %q err=%v, want b", rec.RDESessionID, err)
	}

	// continue + target is contradictory.
	if _, err := resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{continueLatest: true, target: "a"}, nil); err == nil {
		t.Error("expected error combining --continue with a target")
	}

	// unknown target.
	if _, err := resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{target: "nope"}, nil); err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestResolveResumeRecordNoSessions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := resolveResumeRecord(context.Background(), &cobra.Command{}, nil, "/empty/repo", resumeOptions{continueLatest: true}, nil); !errors.Is(err, errNoSessions) {
		t.Errorf("continue with no sessions: err=%v, want errNoSessions", err)
	}
}

// TestResolveResumeRecordSkipsRejectedCandidates covers the fix for a
// legacy-path-only record: localsession.Remove never actually deletes it (it
// only ever writes to the current location), so without skip, --continue would
// keep resolving to the same record on every retry after handleUnresumable
// reports it as gone. skip is how runResume's loop makes progress instead.
func TestResolveResumeRecordSkipsRejectedCandidates(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"
	mustSave(t, localsession.Record{RDESessionID: "a", WorkspaceID: "ws", Name: "claude-a", RepoPath: repoPath})
	time.Sleep(5 * time.Millisecond)
	mustSave(t, localsession.Record{RDESessionID: "b", WorkspaceID: "ws", Name: "claude-b", RepoPath: repoPath})

	cmd := &cobra.Command{}

	// b is newest, but already rejected this run → falls through to a.
	rec, err := resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{continueLatest: true}, map[string]bool{"b": true})
	if err != nil || rec.RDESessionID != "a" {
		t.Errorf("continue skipping b = %q err=%v, want a", rec.RDESessionID, err)
	}

	// Every candidate rejected → a clean error, not the same record forever
	// (what would otherwise happen if a stale record could never be removed).
	if _, err := resolveResumeRecord(context.Background(), cmd, nil, repoPath, resumeOptions{continueLatest: true}, map[string]bool{"a": true, "b": true}); !errors.Is(err, errNoSessions) {
		t.Errorf("continue with everything skipped: err=%v, want errNoSessions", err)
	}
}

func TestHandleUnresumable(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"
	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	log := newStepLogger(cmd)

	// Picker / --continue (no explicit target): the stale record is removed and
	// the caller is told to try the next candidate.
	mustSave(t, localsession.Record{RDESessionID: "a", WorkspaceID: "ws", Name: "claude-a", RepoPath: repoPath})
	retry, err := handleUnresumable(log, repoPath, localsession.Record{RDESessionID: "a", RepoPath: repoPath, Name: "claude-a"}, "no longer exists", resumeOptions{})
	if !retry || err != nil {
		t.Errorf("picker path: retry=%v err=%v, want true,nil", retry, err)
	}
	if _, lerr := localsession.Load(repoPath, "a"); lerr == nil {
		t.Error("record should have been removed")
	}

	// Explicit SESSION_ID: removed and reported as an error (no fall-through).
	mustSave(t, localsession.Record{RDESessionID: "b", WorkspaceID: "ws", Name: "claude-b", RepoPath: repoPath})
	retry, err = handleUnresumable(log, repoPath, localsession.Record{RDESessionID: "b", RepoPath: repoPath, Name: "claude-b"}, "can't be restored", resumeOptions{target: "b"})
	if retry || err == nil {
		t.Errorf("target path: retry=%v err=%v, want false,non-nil", retry, err)
	}
	if _, lerr := localsession.Load(repoPath, "b"); lerr == nil {
		t.Error("record should have been removed")
	}
}

func mustSave(t *testing.T, rec localsession.Record) {
	t.Helper()
	if err := localsession.Save(rec); err != nil {
		t.Fatalf("Save %s: %v", rec.RDESessionID, err)
	}
}

func TestStatusLabel(t *testing.T) {
	for _, tc := range []struct {
		rs   recordStatus
		want string
	}{
		{recordStatus{status: "running", resumable: true}, "running"},
		{recordStatus{status: "terminated", resumable: true}, "terminated"},
		{recordStatus{status: "terminated", resumable: false}, "terminated · unrestorable"},
		{recordStatus{status: "deleted", resumable: false}, "deleted"},
		{recordStatus{}, "status unknown"},
	} {
		if got := statusLabel(tc.rs); got != tc.want {
			t.Errorf("statusLabel(%+v) = %q, want %q", tc.rs, got, tc.want)
		}
	}
}

func TestStatusTone(t *testing.T) {
	for _, tc := range []struct {
		rs   recordStatus
		want picker.Tone
	}{
		{recordStatus{status: "running", resumable: true}, picker.ToneSuccess},
		{recordStatus{status: "failed", resumable: true}, picker.ToneDanger},
		{recordStatus{status: "deleted"}, picker.ToneDanger},
		{recordStatus{status: "terminated", resumable: false}, picker.ToneWarn},
		{recordStatus{status: "terminated", resumable: true}, picker.ToneDim},
		{recordStatus{}, picker.ToneDim},
	} {
		if got := statusTone(tc.rs); got != tc.want {
			t.Errorf("statusTone(%+v) = %v, want %v", tc.rs, got, tc.want)
		}
	}
}

// fakeGetter returns a canned session/error per session ID for status tests.
type fakeGetter struct {
	sessions map[string]internalrde.Session
	errs     map[string]error
}

func (f fakeGetter) GetSession(_ context.Context, _, sessionID string) (internalrde.Session, error) {
	if err := f.errs[sessionID]; err != nil {
		return internalrde.Session{}, err
	}
	return f.sessions[sessionID], nil
}

func TestFetchStatuses(t *testing.T) {
	recs := []localsession.Record{
		{RDESessionID: "run", WorkspaceID: "ws"},
		{RDESessionID: "gone", WorkspaceID: "ws"},
		{RDESessionID: "deaddisk", WorkspaceID: "ws"},
		{RDESessionID: "flaky", WorkspaceID: "ws"},
	}
	getter := fakeGetter{
		sessions: map[string]internalrde.Session{
			"run":      {ID: "run", Status: "running"},
			"deaddisk": {ID: "deaddisk", Status: "terminated", PersistentDiskStatus: internalrde.DiskStatusUnavailable},
		},
		errs: map[string]error{
			"gone":  &rdeapi.APIError{StatusCode: http.StatusNotFound},
			"flaky": errors.New("connection reset"),
		},
	}

	got := fetchStatuses(context.Background(), getter, recs)
	if len(got) != len(recs) {
		t.Fatalf("got %d statuses, want %d", len(got), len(recs))
	}
	// Index-aligned with recs.
	if got[0].status != "running" || !got[0].resumable {
		t.Errorf("run: %+v", got[0])
	}
	if got[1].status != "deleted" || got[1].resumable {
		t.Errorf("gone (404): %+v, want deleted/not-resumable", got[1])
	}
	if got[2].status != "terminated" || got[2].resumable {
		t.Errorf("deaddisk: %+v, want terminated/not-resumable", got[2])
	}
	if got[3].status != "" || !got[3].resumable {
		t.Errorf("flaky (transient err): %+v, want unknown/assumed-resumable", got[3])
	}
}
