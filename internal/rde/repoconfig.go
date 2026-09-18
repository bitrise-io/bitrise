package rde

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// repoConfigDirName/repoConfigFileName name the repo-level RDE dotfile,
// .bitrise/rde.yml, looked up in the working directory and its ancestors.
const (
	repoConfigDirName  = ".bitrise"
	repoConfigFileName = "rde.yml"
)

// RepoConfig is the parsed .bitrise/rde.yml — the repo-level RDE dotfile.
// This is the initial schema; new sections are additive, and unknown keys
// are ignored so an older CLI keeps working against a newer file.
type RepoConfig struct {
	Exec RepoExecConfig `yaml:"exec"`
}

// RepoExecConfig configures `rde session exec` for a repo.
type RepoExecConfig struct {
	// Env lists environment variables forwarded to every exec: NAME
	// (forward the local value; skipped with a warning when unset locally)
	// or NAME=VALUE (a literal).
	Env []string `yaml:"env"`
}

// LoadRepoConfig searches the current working directory and its ancestors
// for the repo-level RDE dotfile (.bitrise/rde.yml). Returns the parsed
// config, the path of the file that was used (empty if none found), and any
// read/parse error. A missing file at all levels is not an error. (Mirrors
// config.LoadDir's discovery for .bitrise-cli.yml: first hit wins, the walk
// goes to the filesystem root.)
func LoadRepoConfig() (RepoConfig, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return RepoConfig{}, "", fmt.Errorf("get working dir: %w", err)
	}
	return loadRepoConfigFrom(cwd)
}

func loadRepoConfigFrom(start string) (RepoConfig, string, error) {
	for dir := start; ; {
		p := filepath.Join(dir, repoConfigDirName, repoConfigFileName)
		data, err := os.ReadFile(p) // p is an ancestor-directory config path, not user input
		if err == nil {
			var c RepoConfig
			if err := yaml.Unmarshal(data, &c); err != nil {
				return RepoConfig{}, "", fmt.Errorf("parse %s: %w", p, err)
			}
			return c, p, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return RepoConfig{}, "", fmt.Errorf("read %s: %w", p, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return RepoConfig{}, "", nil // reached filesystem root
		}
		dir = parent
	}
}
