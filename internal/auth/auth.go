// Package auth persists and reads the Bitrise access token.
//
// Storage: YAML at $XDG_CONFIG_HOME/bitrise/cli/auth.yaml, or
// ~/.config/bitrise/cli/auth.yaml when XDG_CONFIG_HOME is unset. Credentials
// live in their own file, separate from preferences in config.yml, at 0600
// permissions. OS-keychain integration is intentionally deferred.
//
// Load also has a read-only fallback to the predecessor standalone CLI's
// auth.yaml (one directory segment up — see config.PredecessorDir — same
// filename), so an existing install of that CLI isn't silently logged out
// by the merge. That file is never written to, but Clear removes it too, so
// logout can't appear to succeed while a stale token remains discoverable.
//
// The Bitrise API accepts both Personal Access Tokens (user-scoped) and
// Workspace API Tokens (workspace-scoped); they have identical wire format
// and authenticate the same way, so this package treats them as a single
// opaque token. If/when cross-workspace warnings become useful, a "type"
// field can be added back without breaking existing auth.yaml files.
package auth

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// EnvToken overrides the stored access token; it takes precedence over auth.yaml.
const EnvToken = "BITRISE_TOKEN"

// Auth is the on-disk shape of auth.yaml.
//
// Token is the working credential read by every command — a Personal Access
// Token, whether pasted, minted via email login, or obtained through the OAuth
// flow. The remaining fields are populated only by the OAuth flow
// (`auth login --oauth`) to power transparent token refresh; a pasted or
// email-login auth.yaml carries just Token, and such "manual" tokens are never
// refreshed. All OAuth fields are omitempty so manual files stay minimal and
// older files (Token only) keep loading unchanged.
type Auth struct {
	Token        string    `yaml:"token,omitempty"`
	TokenExpiry  time.Time `yaml:"token_expiry,omitempty"`
	JWT          string    `yaml:"jwt,omitempty"`
	JWTExpiry    time.Time `yaml:"jwt_expiry,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
}

// IsOAuthManaged reports whether this token was obtained through the OAuth
// flow and can therefore be refreshed. The refresh token is the distinguishing
// marker: only the OAuth path persists one. Pasted/email-login tokens have an
// empty RefreshToken and are used verbatim.
func (a Auth) IsOAuthManaged() bool {
	return a.RefreshToken != ""
}

func TokenType(token string) string {
	switch {
	case strings.HasPrefix(token, "bitpat_"):
		return "PAT"
	case strings.HasPrefix(token, "bitwat_"):
		return "WAT"
	default:
		return "unknown"
	}
}

func Path() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "auth.yaml"), nil
}

func predecessorPath() (string, error) {
	dir, err := config.PredecessorDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "auth.yaml"), nil
}

// ActivePath returns the auth file Load() actually reads from right now:
// the predecessor CLI's auth.yaml while that fallback is live, or the
// current path once anything has written it (or if neither exists yet).
// Used by `bitrise auth status` so its reported path matches reality —
// Path() alone would name a file that doesn't exist during the fallback
// window.
func ActivePath() (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	if _, statErr := os.Stat(p); errors.Is(statErr, fs.ErrNotExist) {
		pp, err := predecessorPath()
		if err != nil {
			return "", err
		}
		if _, statErr := os.Stat(pp); statErr == nil {
			return pp, nil
		}
	}
	return p, nil
}

// Load reads the auth file. A missing file returns the zero Auth so
// first-time users don't see failures. When it's absent, Load falls back to
// reading the predecessor CLI's auth.yaml (see predecessorPath), never
// writing to it. No key aliasing is needed here, unlike internal/config.Load:
// the shared keys are spelled the same on both sides. The predecessor also
// writes refresh_token_expiry, which this Auth has no field for and so drops
// on read — harmless, since the refresh ladder in internal/oauth tries the
// refresh token and lets the server reject an expired one.
func Load() (Auth, error) {
	p, err := Path()
	if err != nil {
		return Auth{}, err
	}
	pp, err := predecessorPath()
	if err != nil {
		return Auth{}, err
	}
	a, usedFallback, err := config.LoadYAMLWithFallback[Auth](p, pp)
	if err != nil {
		return Auth{}, err
	}
	if usedFallback {
		announceFallback(pp)
	}
	return a, nil
}

var fallbackAnnounceOnce sync.Once

// fallbackWriter is a var, not a bare os.Stderr use, so tests can capture
// the one-time announcement. Kept separate from internal/config's own
// instance of this — sharing one Once would mean whichever file falls back
// first silently suppresses the announcement for the other.
var fallbackWriter io.Writer = os.Stderr

func announceFallback(path string) {
	fallbackAnnounceOnce.Do(func() {
		fmt.Fprintf(fallbackWriter, "Using the previous bitrise-cli's credentials file at %s (read-only; log in again with this CLI to migrate it)\n", path)
	})
}

// Save atomically writes a to disk with 0600 permissions, creating the
// parent directory (0700) if needed. auth.yaml intentionally persists OAuth
// material (PAT/JWT/refresh token) alongside the token — that's the file's
// purpose.
func Save(a Auth) error {
	if a.Token == "" {
		return fmt.Errorf("refusing to save auth with empty token")
	}
	p, err := Path()
	if err != nil {
		return err
	}
	return config.SaveYAML(p, a)
}

// Clear removes the auth file and the predecessor CLI's auth.yaml if it
// exists. A non-existent file is not an error. Removing both matters: with
// only the current path cleared, a later Load would fall back to the
// predecessor file and report the user still logged in — and this doesn't
// revoke the token server-side either, so leaving that file behind would
// keep it live.
func Clear() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", p, err)
	}
	pp, err := predecessorPath()
	if err != nil {
		return err
	}
	if err := os.Remove(pp); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", pp, err)
	}
	return nil
}
