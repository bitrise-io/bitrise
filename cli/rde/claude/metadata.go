package claude

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// generateClaudeSessionID returns a random UUIDv4 to pass to
// `claude --session-id`. Generating it ourselves means we know the in-session
// Claude session ID up front: we can store it immediately, find its transcript
// to read the AI title, and resume it reliably with `claude --resume <id>`.
// Built from crypto/rand to avoid a UUID dependency.
func generateClaudeSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate claude session id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// newDescriber returns a function that builds the session description from the
// local repo context. It's a function (not a fixed string) because the PR can
// appear after the session starts, so the metadata monitor re-evaluates it
// each tick.
func newDescriber(repoSlug, branch string) func(context.Context) string {
	return func(ctx context.Context) string {
		return buildDescription(repoSlug, branch, ghPRURL(ctx, branch))
	}
}

// buildDescription assembles the session description: "owner/repo @ branch"
// with the pull-request URL on its own line. Empty parts are omitted. The URL
// is kept bare so URL-detecting UIs render it as a link.
func buildDescription(repoSlug, branch, prURL string) string {
	desc := repoSlug
	if branch != "" {
		if desc != "" {
			desc += " @ " + branch
		} else {
			desc = branch
		}
	}
	if prURL != "" {
		if desc != "" {
			desc += "\n" + prURL
		} else {
			desc = prURL
		}
	}
	return strings.TrimSpace(desc)
}

// ghPRURL returns the URL of the open pull request for branch, best-effort via
// the `gh` CLI. Returns "" when gh isn't installed, isn't authenticated, or the
// branch has no associated PR.
func ghPRURL(ctx context.Context, branch string) string {
	if branch == "" {
		return ""
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return ""
	}
	// branch comes from the local repo's checked-out HEAD, passed as its own argv element — no shell, no injection
	out, err := exec.CommandContext(ctx, "gh", "pr", "view", branch, "--json", "url", "-q", ".url").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// repoRootPath returns the absolute path of the current repo's root (the
// project key for the local session store). Falls back to the working
// directory, then "".
func repoRootPath(ctx context.Context) string {
	if out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output(); err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			return p
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return ""
}

// repoSlugFromURL extracts "owner/repo" from a clone URL (ssh, scp-like, or
// https), trimming any ".git" suffix. Returns "" when it can't parse one.
func repoSlugFromURL(raw string) string {
	s := strings.TrimSpace(raw)
	scheme := false
	for _, p := range []string{"ssh://", "https://", "http://", "git://"} {
		if rest, ok := strings.CutPrefix(s, p); ok {
			s = rest
			scheme = true
			break
		}
	}
	if _, after, ok := strings.Cut(s, "@"); ok {
		s = after
	}
	var path string
	if scheme {
		// host[:port]/owner/repo
		_, p, ok := strings.Cut(s, "/")
		if !ok {
			return ""
		}
		path = p
	} else {
		// scp-like: host:owner/repo
		_, p, ok := strings.Cut(s, ":")
		if !ok {
			return ""
		}
		path = p
	}
	path = strings.TrimSuffix(path, ".git")
	return strings.Trim(path, "/")
}
