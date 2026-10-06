package update

import (
	"os"
	"path/filepath"
	"time"

	internalconfig "github.com/bitrise-io/bitrise/v3/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	stateFileName = "update-check.yml"
	checkInterval = 24 * time.Hour
)

// State is the content of the update check state file.
type State struct {
	CLI CheckState `yaml:"cli"`
}

type CheckState struct {
	CheckedAt time.Time `yaml:"checked_at"`
	// Once an update has been installed, the findings below describe a version that
	// is no longer running.
	CheckedVersion string   `yaml:"checked_version"`
	Versions       Versions `yaml:"versions,omitempty"`
}

// A missing or unreadable file reads as an empty state, which is due for a
// check: this cache must never turn into an error the user sees.
func LoadState() State {
	path, err := statePath()
	if err != nil {
		return State{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return State{}
	}
	return s
}

func (s CheckState) NeedsCheck(now time.Time, currentVersion string) bool {
	return s.CheckedVersion != currentVersion || now.Sub(s.CheckedAt) >= checkInterval
}

// Call it only after a check succeeded, so a failed request is retried on the
// next run instead of silencing the notice for a day.
func SaveState(s State) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	return internalconfig.SaveYAML(path, s)
}

func statePath() (string, error) {
	dir, err := internalconfig.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}
