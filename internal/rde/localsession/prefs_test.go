package localsession

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestPrefsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"

	if err := SavePrefs(repoPath, Prefs{Stack: "osx-xcode-16.0.x-edge", MachineType: "g2.mac.m2pro.4c-6g"}); err != nil {
		t.Fatalf("SavePrefs: %v", err)
	}
	got, err := LoadPrefs(repoPath)
	if err != nil {
		t.Fatalf("LoadPrefs: %v", err)
	}
	if got.Stack != "osx-xcode-16.0.x-edge" || got.MachineType != "g2.mac.m2pro.4c-6g" {
		t.Errorf("round-trip = %+v, want osx-xcode-16.0.x-edge / g2.mac.m2pro.4c-6g", got)
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be stamped on save")
	}
}

func TestPrefsOverwrite(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"

	if err := SavePrefs(repoPath, Prefs{Stack: "a", MachineType: "x"}); err != nil {
		t.Fatalf("SavePrefs first: %v", err)
	}
	if err := SavePrefs(repoPath, Prefs{Stack: "b", MachineType: "y"}); err != nil {
		t.Fatalf("SavePrefs second: %v", err)
	}
	got, err := LoadPrefs(repoPath)
	if err != nil {
		t.Fatalf("LoadPrefs: %v", err)
	}
	if got.Stack != "b" || got.MachineType != "y" {
		t.Errorf("after overwrite = %+v, want b / y", got)
	}
}

func TestLoadPrefsMissingIsZeroValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := LoadPrefs("/never/saved")
	if err != nil {
		t.Fatalf("LoadPrefs missing: %v", err)
	}
	if got.Stack != "" || got.MachineType != "" {
		t.Errorf("missing prefs = %+v, want zero value", got)
	}
}

// TestPrefsNotListedAsSession guards against the prefs file (which lives in the
// same per-project directory and also ends in .json) being parsed as a phantom
// session by the resume readers.
func TestPrefsNotListedAsSession(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"

	if err := SavePrefs(repoPath, Prefs{Stack: "a", MachineType: "x"}); err != nil {
		t.Fatalf("SavePrefs: %v", err)
	}
	if err := Save(Record{RDESessionID: "sess-1", RepoPath: repoPath, Name: "claude-1"}); err != nil {
		t.Fatalf("Save record: %v", err)
	}

	recs, err := ListByProject(repoPath)
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(recs) != 1 || recs[0].RDESessionID != "sess-1" {
		t.Fatalf("ListByProject = %+v, want exactly the real session sess-1", recs)
	}

	latest, ok, err := Latest(repoPath)
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if !ok || latest.RDESessionID != "sess-1" {
		t.Errorf("Latest = %+v (ok=%v), want sess-1", latest, ok)
	}
}

func TestSavePrefsValidation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := SavePrefs("", Prefs{Stack: "a"}); err == nil {
		t.Error("SavePrefs with empty repo path should error")
	}
}

func TestPrefsFilePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix perms")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"

	if err := SavePrefs(repoPath, Prefs{Stack: "a", MachineType: "x"}); err != nil {
		t.Fatalf("SavePrefs: %v", err)
	}
	dir, _ := projectDir(repoPath)
	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if di.Mode().Perm() != 0o700 {
		t.Errorf("dir perm = %v, want 0700", di.Mode().Perm())
	}
	fi, err := os.Stat(filepath.Join(dir, prefsFileName))
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("file perm = %v, want 0600", fi.Mode().Perm())
	}
}

// TestLoadPrefsFallsBackToLegacyLocation is the prefs.go counterpart to
// store_test.go's TestLoadFallsBackToLegacyLocation: an existing user's saved
// stack/machine-type choice must still be found at the pre-switch location.
func TestLoadPrefsFallsBackToLegacyLocation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repoPath := "/work/repo"

	legacyDir, err := legacyProjectDir(repoPath)
	if err != nil {
		t.Fatalf("legacyProjectDir: %v", err)
	}
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	legacyPrefs := Prefs{Stack: "legacy-stack", MachineType: "legacy-machine", UpdatedAt: time.Now().UTC()}
	data, err := json.MarshalIndent(legacyPrefs, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy prefs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, prefsFileName), data, 0o600); err != nil {
		t.Fatalf("write legacy prefs: %v", err)
	}

	got, err := LoadPrefs(repoPath)
	if err != nil {
		t.Fatalf("LoadPrefs did not fall back to legacy location: %v", err)
	}
	if got.Stack != "legacy-stack" || got.MachineType != "legacy-machine" {
		t.Errorf("LoadPrefs fallback = %+v, want legacy-stack / legacy-machine", got)
	}
}
