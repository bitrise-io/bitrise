// Package auth persists and reads the Bitrise access token.
//
// Storage: YAML at $XDG_CONFIG_HOME/bitrise/cli/auth.yaml, falling back to
// ~/.config/bitrise/cli/auth.yaml. Credentials live in their own file,
// separate from preferences in config.yml, at 0600 permissions. OS-keychain
// integration is intentionally deferred.
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
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	// RefreshTokenExpiry is when the refresh token itself expires; past this,
	// the OAuth ladder can no longer recover and the user must re-run login.
	RefreshTokenExpiry time.Time `yaml:"refresh_token_expiry,omitempty"`
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

// Load reads the auth file. A missing file returns the zero Auth so
// first-time users don't see failures.
func Load() (Auth, error) {
	p, err := Path()
	if err != nil {
		return Auth{}, err
	}
	return config.LoadYAML[Auth](p)
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

// Clear removes the auth file. A non-existent file is not an error.
func Clear() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", p, err)
	}
	return nil
}
