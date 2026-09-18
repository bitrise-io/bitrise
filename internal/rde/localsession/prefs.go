package localsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// prefsFileName is the per-project file holding the last stack / machine type
// chosen for `rde claude`, so the next run in the same repo can preselect them.
// It sits at the per-project root <config-dir>/rde/projects/<key>/prefs.json,
// beside the sessions/ subdirectory that holds the session records.
const prefsFileName = "prefs.json"

// Prefs is the remembered `rde claude` selection for one local repo. Values are
// the stack ID and machine type NAME (what CreateSession takes on the wire),
// not cluster-specific IDs.
type Prefs struct {
	Stack       string    `json:"stack,omitempty"`
	MachineType string    `json:"machine_type,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// LoadPrefs returns the remembered selection for the given local repo path. It
// checks the current location first, falling back to the legacy pre-switch
// location on a miss (symmetric with Load/ListByProject in store.go), so an
// existing user's saved stack/machine-type choice isn't silently forgotten
// after the binary switch. A missing file (in both locations) is not an
// error — it returns the zero Prefs, signalling "no prior choice" so the
// caller falls back to the backend default / first item.
func LoadPrefs(repoPath string) (Prefs, error) {
	p, err := readPrefs(repoPath, projectDir)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return Prefs{}, err
	}
	p, err = readPrefs(repoPath, legacyProjectDir)
	if errors.Is(err, fs.ErrNotExist) {
		return Prefs{}, nil
	}
	return p, err
}

func readPrefs(repoPath string, dirFn func(string) (string, error)) (Prefs, error) {
	dir, err := dirFn(repoPath)
	if err != nil {
		return Prefs{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, prefsFileName))
	if errors.Is(err, fs.ErrNotExist) {
		return Prefs{}, fs.ErrNotExist
	}
	if err != nil {
		return Prefs{}, fmt.Errorf("read project prefs: %w", err)
	}
	var p Prefs
	if err := json.Unmarshal(data, &p); err != nil {
		// A corrupt prefs file shouldn't block a run — treat it as "no prior
		// choice" rather than failing the command.
		return Prefs{}, nil
	}
	return p, nil
}

// SavePrefs writes (or overwrites) the remembered selection for the given local
// repo path. It stamps UpdatedAt and writes atomically. Directories are 0700,
// the file 0600 — matching the sibling session records and auth.yaml/config.yaml.
// Always writes to the current (non-legacy) location.
func SavePrefs(repoPath string, p Prefs) error {
	if repoPath == "" {
		return errors.New("prefs have no repo path")
	}
	p.UpdatedAt = time.Now().UTC()

	dir, err := projectDir(repoPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create project prefs dir: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project prefs: %w", err)
	}
	data = append(data, '\n')

	if err := writeFileAtomic(dir, prefsFileName, data); err != nil {
		return fmt.Errorf("save project prefs: %w", err)
	}
	return nil
}
