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

// State is the result of the last successful release check. It has its own file
// rather than a place in config.yml, so the config schema stays stable and a
// state file that fails to load is free to be ignored.
type State struct {
	CheckedAt time.Time `yaml:"checked_at"`
	// CheckedVersion is the CLI version the check ran for. Once the CLI has been
	// updated the findings below describe a build that is no longer running.
	CheckedVersion string   `yaml:"checked_version"`
	Versions       Versions `yaml:"versions,omitempty"`
}

// LoadState returns the state of the last check. A missing or unreadable file
// reads as an empty state, which is due for a check: this cache must never turn
// into an error the user sees.
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

// NeedsCheck reports whether the releases have to be asked for again.
func (s State) NeedsCheck(now time.Time, currentVersion string) bool {
	return s.CheckedVersion != currentVersion || now.Sub(s.CheckedAt) >= checkInterval
}

// SaveState writes the state atomically, so an exit mid-write cannot leave half
// a file behind. Call it only after a check succeeded, so a failed request is
// retried on the next run instead of silencing the notice for a day.
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
