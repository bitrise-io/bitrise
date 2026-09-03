package claude

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

// Claude Code auth env-var names — also the saved-input keys the RDE control
// plane uses for a configured Subscription Token / API Key.
const (
	envOAuthToken = "CLAUDE_CODE_OAUTH_TOKEN"
	envAPIKey     = "ANTHROPIC_API_KEY"
)

// sourceSetupToken labels a credential freshly minted by `claude setup-token`.
const sourceSetupToken = "claude setup-token"

// forwardCred is a local Claude Code credential to save on the control plane.
type forwardCred struct {
	EnvVar string // saved-input key (CLAUDE_CODE_OAUTH_TOKEN / ANTHROPIC_API_KEY)
	Value  string // the token/key value
	Source string // short human-readable origin, for logging
}

// controlPlaneHasClaudeToken reports whether the user has a Claude Code token
// configured on the control plane — i.e. a saved input keyed
// CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_API_KEY. Saved inputs are user-scoped,
// so no workspace is needed.
func controlPlaneHasClaudeToken(ctx context.Context, svc *internalrde.Service) (bool, error) {
	inputs, err := svc.ListSavedInputs(ctx)
	if err != nil {
		return false, err
	}
	for _, in := range inputs {
		if in.Key == envOAuthToken || in.Key == envAPIKey {
			return true, nil
		}
	}
	return false, nil
}

// existingLocalCredential resolves a Claude Code credential already present on
// the machine, without any interaction. Precedence mirrors how `claude` itself
// resolves auth: an explicit env var wins, then the credentials file. Returns
// ok=false when nothing is found (the caller may then mint one).
func existingLocalCredential() (forwardCred, bool) {
	if v := os.Getenv(envOAuthToken); v != "" {
		return forwardCred{EnvVar: envOAuthToken, Value: v, Source: "$" + envOAuthToken}, true
	}
	if v := os.Getenv(envAPIKey); v != "" {
		return forwardCred{EnvVar: envAPIKey, Value: v, Source: "$" + envAPIKey}, true
	}
	if v, ok := credentialsFileToken(); ok {
		return forwardCred{EnvVar: envOAuthToken, Value: v, Source: credentialsFilePath()}, true
	}
	return forwardCred{}, false
}

// mintSetupToken runs `claude setup-token` to obtain a long-lived OAuth token.
// It's interactive (browser auth): stdin/stderr stay wired to the terminal so
// the user can complete the flow; the token is read from stdout.
func mintSetupToken(ctx context.Context) (forwardCred, bool) {
	c := exec.CommandContext(ctx, "claude", "setup-token")
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	out, err := c.Output()
	if err != nil {
		return forwardCred{}, false
	}
	token := extractToken(string(out))
	if token == "" {
		return forwardCred{}, false
	}
	return forwardCred{EnvVar: envOAuthToken, Value: token, Source: sourceSetupToken}, true
}

var (
	// ansiSeqRe matches the terminal escape sequences (CSI and OSC) that
	// `claude setup-token` interleaves with its output.
	ansiSeqRe = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]|\x1b\][^\x07]*\x07`)
	// tokenRe matches an Anthropic token by its stable prefix.
	tokenRe = regexp.MustCompile(`sk-ant-[A-Za-z0-9._-]{10,}`)
	// fallbackTokenRe accepts a lone token-shaped line if the prefix ever
	// changes — strict enough to reject stray escape-sequence fragments.
	fallbackTokenRe = regexp.MustCompile(`^[A-Za-z0-9._-]{20,}$`)
)

// extractToken pulls the OAuth token out of `claude setup-token` output, which
// interleaves the token with terminal escape sequences (so the naive "last
// line" is often a terminal-restore sequence like ESC[?2004l). It strips ANSI
// and control sequences, then matches the token by its sk-ant- prefix, falling
// back to a single token-shaped line.
func extractToken(out string) string {
	clean := ansiSeqRe.ReplaceAllString(out, "")
	clean = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if r < 0x20 || r == 0x7f { // drop remaining control chars (incl. lone ESC)
			return -1
		}
		return r
	}, clean)

	if m := tokenRe.FindString(clean); m != "" {
		return m
	}
	if line := lastNonEmptyLine(clean); fallbackTokenRe.MatchString(line) {
		return line
	}
	return ""
}

// lastNonEmptyLine returns the last non-blank, trimmed line of s.
func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}

// credentialsFilePath returns the path to Claude Code's credentials file,
// honoring $CLAUDE_CONFIG_DIR and falling back to ~/.claude.
func credentialsFilePath() string {
	dir := os.Getenv("CLAUDE_CONFIG_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".claude")
	}
	return filepath.Join(dir, ".credentials.json")
}

func credentialsFileToken() (string, bool) {
	p := credentialsFilePath()
	if p == "" {
		return "", false
	}
	// p is the user's own Claude Code credentials file under $HOME/$CLAUDE_CONFIG_DIR
	data, err := os.ReadFile(p)
	if err != nil {
		return "", false
	}
	return parseClaudeAccessToken(data)
}

// parseClaudeAccessToken extracts the OAuth access token from a Claude Code
// credentials blob ({"claudeAiOauth":{"accessToken":…}}). If the blob isn't
// that JSON shape, the trimmed raw bytes are treated as a bare token.
func parseClaudeAccessToken(data []byte) (string, bool) {
	var creds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &creds); err == nil && creds.ClaudeAiOauth.AccessToken != "" {
		return creds.ClaudeAiOauth.AccessToken, true
	}
	if raw := strings.TrimSpace(string(data)); raw != "" && !strings.HasPrefix(raw, "{") {
		return raw, true
	}
	return "", false
}
