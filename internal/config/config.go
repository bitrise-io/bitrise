// Package config persists and reads the CLI's layered settings: a global
// config file and a per-directory override file, both layered underneath the
// pre-existing ~/.bitrise/config.json store, which stays authoritative when
// present (see resolve.go).
//
// Storage: YAML at $XDG_CONFIG_HOME/bitrise/cli/config.yml, or
// ~/.config/bitrise/cli/config.yml when XDG_CONFIG_HOME is unset.
// Written with 0600 permissions.
//
// Load also has a read-only fallback to the predecessor standalone CLI's
// config.yaml — one directory segment up (no "cli"), and a .yaml extension
// — so an existing install of that CLI isn't silently logged out of its
// settings by the merge. That file is never written to; see
// PredecessorDir and LoadYAMLWithFallback. (This is a different fallback
// from the ~/.bitrise/config.json layering above: that one is an
// always-consulted, higher-precedence layer; this one only fires once the
// new config.yml doesn't exist yet.)
package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk shape. SetupVersion, LastCLIUpdateCheck, and
// LastPluginUpdateChecks mirror configs.ConfigModel exactly, so the legacy
// ~/.bitrise/config.json can act as a layer in Resolve (see resolve.go) with
// the same fields to compare against. Newer fields like APIBaseURL and
// WebBaseURL have no legacy counterpart — the legacy layer simply leaves
// them zero.
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
	WebBaseURL             string               `yaml:"web_base_url,omitempty"`
	RDEAPIBaseURL          string               `yaml:"rde_api_base_url,omitempty"`
	AppID                  string               `yaml:"app_id,omitempty"`
	DefaultWorkspaceID     string               `yaml:"default_workspace_id,omitempty"`
	Output                 string               `yaml:"output,omitempty"`
	Theme                  string               `yaml:"theme,omitempty"`
}

// DirFileName is the file looked up in the working directory and its
// ancestors to provide per-project overrides — above the global file, below
// env vars/flags.
const DirFileName = ".bitrise-cli.yml"

// Dir returns the absolute path to the bitrise CLI config directory — the
// parent of the global config file (see the package doc for the XDG
// fallback rule).
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

// PredecessorDir returns the config directory of the predecessor standalone
// CLI that this repo merged in — one path segment above Dir(), which this
// repo added. Read-only fallback target; nothing here is ever written back
// to it, since that CLI is still in use for a while after this merge.
func PredecessorDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Dir(dir), nil
}

func predecessorConfigPath() (string, error) {
	dir, err := PredecessorDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil // note: .yaml, not .yml
}

// predecessorConfig is a superset of Config: every current field (so
// unmarshaling the current config.yml through it loses nothing), plus the
// predecessor CLI's pre-rename key spellings (app_slug,
// default_workspace_slug) as fallbacks, so a fallen-back-to file doesn't
// silently drop those values either.
type predecessorConfig struct {
	SetupVersion           string               `yaml:"setup_version"`
	LastCLIUpdateCheck     time.Time            `yaml:"last_cli_update_check"`
	LastPluginUpdateChecks map[string]time.Time `yaml:"last_plugin_update_checks"`
	Output                 string               `yaml:"output"`
	AppID                  string               `yaml:"app_id"`
	AppSlugLegacy          string               `yaml:"app_slug"`
	DefaultWorkspaceID     string               `yaml:"default_workspace_id"`
	DefaultWorkspaceSlug   string               `yaml:"default_workspace_slug"`
	APIBaseURL             string               `yaml:"api_base_url"`
	RDEAPIBaseURL          string               `yaml:"rde_api_base_url"`
	WebBaseURL             string               `yaml:"web_base_url"`
	Theme                  string               `yaml:"theme"`
}

func (p predecessorConfig) toConfig() Config {
	return Config{
		SetupVersion:           p.SetupVersion,
		LastCLIUpdateCheck:     p.LastCLIUpdateCheck,
		LastPluginUpdateChecks: p.LastPluginUpdateChecks,
		Output:                 p.Output,
		AppID:                  FirstNonEmptyString(p.AppID, p.AppSlugLegacy),
		DefaultWorkspaceID:     FirstNonEmptyString(p.DefaultWorkspaceID, p.DefaultWorkspaceSlug),
		APIBaseURL:             p.APIBaseURL,
		RDEAPIBaseURL:          p.RDEAPIBaseURL,
		WebBaseURL:             p.WebBaseURL,
		Theme:                  p.Theme,
	}
}

var fallbackAnnounceOnce sync.Once

// fallbackWriter is a var, not a bare os.Stderr use, so tests can capture
// the one-time announcement.
var fallbackWriter io.Writer = os.Stderr

// announceFallback prints, once per process, that the predecessor CLI's
// config file is being read as a fallback. internal/auth keeps its own
// separate instance of this (own sync.Once) for its own file — sharing one
// Once across both would mean whichever file falls back first silently
// suppresses the announcement for the other.
func announceFallback(what, path string) {
	fallbackAnnounceOnce.Do(func() {
		fmt.Fprintf(fallbackWriter, "Using the previous bitrise-cli's %s at %s (read-only; save with this CLI to migrate it)\n", what, path)
	})
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

// LoadYAMLWithFallback behaves like LoadYAML, but when path is absent it
// reads fallbackPath instead — never writing to it. ok reports whether
// fallbackPath supplied the value, so a caller can announce the fallback.
func LoadYAMLWithFallback[T any](path, fallbackPath string) (v T, ok bool, err error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		data, err = os.ReadFile(fallbackPath)
		if errors.Is(err, fs.ErrNotExist) {
			return v, false, nil
		}
		if err != nil {
			return v, false, fmt.Errorf("read %s: %w", fallbackPath, err)
		}
		if uerr := yaml.Unmarshal(data, &v); uerr != nil {
			return v, false, fmt.Errorf("parse %s: %w", fallbackPath, uerr)
		}
		return v, true, nil
	}
	if err != nil {
		return v, false, fmt.Errorf("read %s: %w", path, err)
	}
	if uerr := yaml.Unmarshal(data, &v); uerr != nil {
		return v, false, fmt.Errorf("parse %s: %w", path, uerr)
	}
	return v, false, nil
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
	return writeAtomic(path, data)
}

// Load reads the global config file. When it's absent, Load falls back to
// reading the predecessor standalone CLI's config.yaml (see PredecessorDir),
// never writing to it, so an existing install of that CLI doesn't silently
// lose app_id, default_workspace_id, theme, output and any base-URL
// overrides to this merge. cli/config/{set,unset}.go are
// Load-mutate-Save, so the fallback has to live here rather than only in
// Resolve: the first `config set` reads the fallback content in full and
// writes it all forward, completing the move on that first write, instead
// of writing a file holding only the one just-set key and losing the rest.
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	pp, err := predecessorConfigPath()
	if err != nil {
		return Config{}, err
	}
	pc, usedFallback, err := LoadYAMLWithFallback[predecessorConfig](p, pp)
	if err != nil {
		return Config{}, err
	}
	if usedFallback {
		announceFallback("config file", pp)
	}
	return pc.toConfig(), nil
}

// ActivePath returns the config file Load() actually reads from right now:
// the predecessor CLI's config.yaml while that fallback is live, or the new
// config.yml once anything has written it (or if neither exists yet).
func ActivePath() (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	if _, statErr := os.Stat(p); errors.Is(statErr, fs.ErrNotExist) {
		pp, err := predecessorConfigPath()
		if err != nil {
			return "", err
		}
		if _, statErr := os.Stat(pp); statErr == nil {
			return pp, nil
		}
	}
	return p, nil
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
		data, err := os.ReadFile(p)
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
	return writeAtomic(p, data)
}

// staleTmpAge bounds how old a *.tmp sibling has to be before writeAtomic
// sweeps it — comfortably longer than any real write, so a concurrent
// writer's own in-flight temp file (same glob pattern, near-zero age) is
// never mistaken for one stranded by a crash.
const staleTmpAge = time.Minute

// writeAtomic writes data to a uniquely-named temp file beside path and
// renames it into place, so two processes writing the same path concurrently
// never race the rename with a shared temp name. It fsyncs before the rename
// so a crash right after can't leave a file that looks written but lost its
// data on an unclean shutdown, and it sweeps *.tmp siblings older than
// staleTmpAge, left behind by a previous call that crashed between create
// and rename — otherwise those leak forever, and for auth.yaml/config.yml
// that means a stranded credential-bearing file.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	if stale, err := filepath.Glob(filepath.Join(dir, base+".*.tmp")); err == nil {
		now := time.Now()
		for _, f := range stale {
			if info, statErr := os.Stat(f); statErr == nil && now.Sub(info.ModTime()) > staleTmpAge {
				_ = os.Remove(f)
			}
		}
	}

	f, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }() // no-op once the rename below succeeds

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync %s: %w", tmp, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("install %s: %w", path, err)
	}
	return nil
}
