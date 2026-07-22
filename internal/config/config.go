// Package config persists and reads the CLI's layered settings: a global
// config file and a per-directory override file, both layered underneath the
// pre-existing ~/.bitrise/config.json store, which stays authoritative when
// present (see resolve.go).
//
// Storage: YAML at $XDG_CONFIG_HOME/bitrise/cli/config.yml, falling back to
// ~/.config/bitrise/cli/config.yml. Written with 0600 permissions.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk shape. SetupVersion, LastCLIUpdateCheck, and
// LastPluginUpdateChecks mirror configs.ConfigModel exactly, so the legacy
// ~/.bitrise/config.json can act as a layer in Resolve (see resolve.go) with
// the same fields to compare against. Newer fields like APIBaseURL have no
// legacy counterpart — the legacy layer simply leaves them zero.
//
// LastCLIUpdateCheck and LastPluginUpdateChecks are timestamps the CLI
// itself writes during normal operation, not user preferences — unusual
// candidates for a hand-edited YAML override — but are included here for
// uniform treatment across all three fields.
type Config struct {
	SetupVersion           string               `yaml:"setup_version,omitempty"`
	LastCLIUpdateCheck     time.Time            `yaml:"last_cli_update_check,omitempty"`
	LastPluginUpdateChecks map[string]time.Time `yaml:"last_plugin_update_checks,omitempty"`
	APIBaseURL             string               `yaml:"api_base_url,omitempty"`
}

// DirFileName is the file looked up in the working directory and its
// ancestors to provide per-project overrides — above the global file, below
// env vars/flags.
const DirFileName = ".bitrise-cli.yml"

// Dir returns the absolute path to the bitrise CLI config directory — the
// parent of the global config file. Honors XDG_CONFIG_HOME, falling back to
// ~/.config/bitrise/cli.
func Dir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate user home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "bitrise", "cli"), nil
}

// Path returns the absolute path to the global config file (whether or not
// it exists).
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// LoadYAML reads and unmarshals the YAML file at path into T. A missing
// file is not an error — it returns the zero T, so callers can treat "not
// yet configured" the same as "empty".
func LoadYAML[T any](path string) (T, error) {
	var v T
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return v, nil
	}
	if err != nil {
		return v, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("parse %s: %w", path, err)
	}
	return v, nil
}

// SaveYAML atomically marshals v to YAML and writes it to path with 0600
// permissions, creating the parent directory (0700) if missing.
func SaveYAML[T any](path string, v T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("install %s: %w", path, err)
	}
	return nil
}

// Load reads the global config file. A missing file is not an error — it
// returns the zero Config so first-time users don't see failures.
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	return LoadYAML[Config](p)
}

// LoadDir searches the current working directory and its ancestors for a
// per-project config file (DirFileName). Returns the parsed config, the
// absolute path of the file that was used (empty if none found), and any
// parse error. A missing file at all levels is not an error.
func LoadDir() (Config, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, "", fmt.Errorf("get working dir: %w", err)
	}
	return loadDirFrom(cwd)
}

func loadDirFrom(start string) (Config, string, error) {
	for dir := start; ; {
		p := filepath.Join(dir, DirFileName)
		data, err := os.ReadFile(p) //nolint:gosec // p is an ancestor-directory config path, not user input
		if err == nil {
			var c Config
			if err := yaml.Unmarshal(data, &c); err != nil {
				return Config{}, "", fmt.Errorf("parse %s: %w", p, err)
			}
			return c, p, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return Config{}, "", fmt.Errorf("read %s: %w", p, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Config{}, "", nil // reached filesystem root
		}
		dir = parent
	}
}

// Save atomically writes c to disk with 0600 permissions. It creates the
// parent directory (0700) if missing.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("install %s: %w", p, err)
	}
	return nil
}
