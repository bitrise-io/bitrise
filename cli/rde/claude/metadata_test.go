package claude

import (
	"regexp"
	"testing"
)

func TestRepoSlugFromURL(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"git@github.com:org/repo.git", "org/repo"},
		{"git@github.com:org/repo", "org/repo"},
		{"https://github.com/org/repo.git", "org/repo"},
		{"https://github.com/org/repo", "org/repo"},
		{"ssh://git@github.com/org/repo.git", "org/repo"},
		{"git@gitlab.com:grp/sub/repo.git", "grp/sub/repo"},
		{"https://user@github.com/org/repo.git", "org/repo"},
		{"not a url", ""},
		{"", ""},
	} {
		if got := repoSlugFromURL(tc.in); got != tc.want {
			t.Errorf("repoSlugFromURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildDescription(t *testing.T) {
	for _, tc := range []struct {
		slug, branch, prURL, want string
	}{
		{"org/repo", "main", "", "org/repo @ main"},
		{"org/repo", "main", "https://example.com/pull/7", "org/repo @ main\nhttps://example.com/pull/7"},
		{"org/repo", "", "", "org/repo"},
		{"", "main", "", "main"},
		{"", "", "https://example.com/pull/7", "https://example.com/pull/7"},
		{"", "", "", ""},
	} {
		if got := buildDescription(tc.slug, tc.branch, tc.prURL); got != tc.want {
			t.Errorf("buildDescription(%q,%q,%q) = %q, want %q", tc.slug, tc.branch, tc.prURL, got, tc.want)
		}
	}
}

func TestGenerateClaudeSessionID(t *testing.T) {
	uuidRe := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	a, err := generateClaudeSessionID()
	if err != nil {
		t.Fatalf("generateClaudeSessionID: %v", err)
	}
	if !uuidRe.MatchString(a) {
		t.Errorf("not a v4 UUID: %q", a)
	}
	b, _ := generateClaudeSessionID()
	if a == b {
		t.Errorf("expected distinct UUIDs, got %q twice", a)
	}
}
