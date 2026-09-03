// Package localsession persists `rde claude` session records locally so they
// can be resumed later (`rde claude --resume` / `--continue`).
//
// Records are grouped by the local repository they were started from, mirroring
// how Claude Code organizes its own transcripts by project:
//
//	<config-dir>/rde/projects/<encoded-repo-path>/sessions/<rde-session-id>.json
//
// where <config-dir> is the bitrise config directory (see internal/config.Dir).
// One file per `rde claude` invocation. The record is written immediately at
// session creation (so an abrupt stop still leaves something resumable) and
// enriched over the session's life by the metadata monitor with the
// AI-generated title and description.
//
// Session records live in their own "sessions" subdirectory so other
// per-project state (e.g. prefs.json, see prefs.go) can sit beside it without
// the readers here having to distinguish them.
package localsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// Record is one resumable `rde claude` session. JSON tags define the on-disk
// shape; this file is read only by the CLI, so the contract is internal.
type Record struct {
	RDESessionID    string    `json:"rde_session_id"`
	WorkspaceID     string    `json:"workspace_id"`
	Name            string    `json:"name"`              // initial generated name (claude-<hex>)
	ClaudeSessionID string    `json:"claude_session_id"` // UUID we pass to `claude --session-id`
	AITitle         string    `json:"ai_title,omitempty"`
	Description     string    `json:"description,omitempty"`
	Repo            string    `json:"repo,omitempty"`      // origin remote URL
	RepoPath        string    `json:"repo_path,omitempty"` // local repo root; the project key
	Branch          string    `json:"branch,omitempty"`
	RemoteRepoDir   string    `json:"remote_repo_dir,omitempty"` // dir Claude runs in on the session
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DisplayName is the human label for a record: the AI-generated title once it
// exists, otherwise the initial generated name.
func (r Record) DisplayName() string {
	if r.AITitle != "" {
		return r.AITitle
	}
	return r.Name
}

// projectsDir is the root holding every project's records.
func projectsDir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "rde", "projects"), nil
}

// legacyProjectsDir is where the predecessor standalone CLI stored these
// records, one directory level above this repo's config.Dir() (which adds a
// "cli" segment the predecessor didn't have). Reads fall back here so an
// existing user's resume history survives the binary switch; nothing ever
// writes here.
func legacyProjectsDir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(dir), "rde", "projects"), nil
}

// projectDir returns the per-project root for the given local repo path. The
// key encodes the path into a single filesystem-safe segment. It holds the
// sessions/ subdirectory and any other per-project state (e.g. prefs.json).
func projectDir(repoPath string) (string, error) {
	root, err := projectsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, projectKey(repoPath)), nil
}

// legacyProjectDir is legacyProjectsDir's per-project counterpart to
// projectDir.
func legacyProjectDir(repoPath string) (string, error) {
	root, err := legacyProjectsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, projectKey(repoPath)), nil
}

// sessionsDir returns the directory holding the session records for the given
// local repo path. Records live under their own subdirectory so sibling
// per-project files (prefs.json) can't be mistaken for sessions.
func sessionsDir(repoPath string) (string, error) {
	dir, err := projectDir(repoPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

// legacySessionsDir is legacySessionsDir's per-project counterpart to
// sessionsDir.
func legacySessionsDir(repoPath string) (string, error) {
	dir, err := legacyProjectDir(repoPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

// projectKey encodes a local repo path into one filesystem-safe segment by
// replacing every character outside [A-Za-z0-9._-] with '-'. The exact scheme
// doesn't need to match Claude Code's — the store only reads its own keys — it
// just has to be stable for a given path.
func projectKey(repoPath string) string {
	var b strings.Builder
	b.Grow(len(repoPath))
	for _, r := range repoPath {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	key := b.String()
	if key == "" {
		key = "-"
	}
	return key
}

// Save writes (or overwrites) the record for its repo + RDE session ID. It
// stamps UpdatedAt (and CreatedAt if unset) and writes atomically so a reader
// never sees a half-written file. Directories are 0700, the file 0600 — the
// record carries no secrets, but it sits alongside auth.yaml/config.yaml and
// follows the same locked-in perms. Always writes to the current (non-legacy)
// location.
func Save(rec Record) error {
	if rec.RDESessionID == "" {
		return errors.New("record has no RDE session ID")
	}
	if rec.RepoPath == "" {
		return errors.New("record has no repo path")
	}
	now := time.Now().UTC()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now

	dir, err := sessionsDir(rec.RepoPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create session store dir: %w", err)
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session record: %w", err)
	}
	data = append(data, '\n')

	if err := writeFileAtomic(dir, rec.RDESessionID+".json", data); err != nil {
		return fmt.Errorf("save session record: %w", err)
	}
	return nil
}

// writeFileAtomic writes data to dir/name via a temp file + rename, so a reader
// never sees a half-written file. The file is 0600; callers create dir 0700.
// The temp file keeps a ".tmp" suffix so it's ignored by the ".json" readers.
func writeFileAtomic(dir, name string, data []byte) error {
	tmp, err := os.CreateTemp(dir, name+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) //nolint:errcheck // best-effort cleanup if rename already moved it
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close() //nolint:errcheck // returning the chmod error; close failure is secondary
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck // returning the write error; close failure is secondary
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("install file: %w", err)
	}
	return nil
}

// Load returns the record for the given repo + RDE session ID. It checks the
// current location first, falling back to the legacy pre-switch location on a
// miss (see legacyProjectsDir). A record found in neither returns
// os.ErrNotExist.
func Load(repoPath, rdeSessionID string) (Record, error) {
	dir, err := sessionsDir(repoPath)
	if err != nil {
		return Record{}, err
	}
	rec, err := readRecord(filepath.Join(dir, rdeSessionID+".json"))
	if err == nil {
		return rec, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return Record{}, err
	}

	legacyDir, lerr := legacySessionsDir(repoPath)
	if lerr != nil {
		return Record{}, err
	}
	return readRecord(filepath.Join(legacyDir, rdeSessionID+".json"))
}

// ListByProject returns every record for the given local repo, newest-updated
// first. It checks the current location first, falling back to the legacy
// pre-switch location only when the current one has nothing (not a merge of
// both — once anything exists at the current location, legacy is no longer
// consulted). A missing sessions directory is not an error. Unparseable files
// are skipped so one corrupt record can't hide the rest.
func ListByProject(repoPath string) ([]Record, error) {
	dir, err := sessionsDir(repoPath)
	if err != nil {
		return nil, err
	}
	recs, err := readSessionsDir(dir)
	if err != nil {
		return nil, err
	}
	if len(recs) > 0 {
		return recs, nil
	}

	legacyDir, err := legacySessionsDir(repoPath)
	if err != nil {
		return nil, err
	}
	return readSessionsDir(legacyDir)
}

func readSessionsDir(dir string) ([]Record, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read session store: %w", err)
	}
	var recs []Record
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		rec, err := readRecord(filepath.Join(dir, e.Name()))
		if err != nil {
			continue // skip corrupt/partial records
		}
		recs = append(recs, rec)
	}
	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].UpdatedAt.After(recs[j].UpdatedAt)
	})
	return recs, nil
}

// Latest returns the most-recently-updated record for the repo, or ok=false
// when none exist.
func Latest(repoPath string) (Record, bool, error) {
	recs, err := ListByProject(repoPath)
	if err != nil {
		return Record{}, false, err
	}
	if len(recs) == 0 {
		return Record{}, false, nil
	}
	return recs[0], true, nil
}

// Remove deletes the record for the given repo + RDE session ID from the
// current (non-legacy) location. A missing record is not an error. A record
// that exists only at the legacy location is untouched — Remove is a write,
// and the legacy fallback is read-only.
func Remove(repoPath, rdeSessionID string) error {
	dir, err := sessionsDir(repoPath)
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(dir, rdeSessionID+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func readRecord(path string) (Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return Record{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return rec, nil
}
