package claude

import (
	"strings"
	"testing"
)

func TestIndentWriter(t *testing.T) {
	for _, tc := range []struct {
		name   string
		writes []string
		want   string
	}{
		{"single line", []string{"hello\n"}, "  hello\n"},
		{"multi line", []string{"a\nb\n"}, "  a\n  b\n"},
		{"split across writes", []string{"hel", "lo\nwor", "ld\n"}, "  hello\n  world\n"},
		{"carriage return progress", []string{"50%\r100%\n"}, "  50%\r  100%\n"},
		{"crlf not double-indented", []string{"a\r\nb\r\n"}, "  a\r\n  b\r\n"},
		{"trailing newline only indents next line lazily", []string{"x\n"}, "  x\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder
			iw := newIndentWriter(&sb)
			for _, w := range tc.writes {
				n, err := iw.Write([]byte(w))
				if err != nil {
					t.Fatalf("Write(%q): %v", w, err)
				}
				if n != len(w) {
					t.Errorf("Write(%q) = %d, want %d", w, n, len(w))
				}
			}
			if got := sb.String(); got != tc.want {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestGitSSHURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"ssh passthrough", "git@github.com:org/repo.git", "git@github.com:org/repo.git"},
		{"ssh scheme passthrough", "ssh://git@github.com/org/repo.git", "ssh://git@github.com/org/repo.git"},
		{"github https rewrite", "https://github.com/org/repo.git", "git@github.com:org/repo.git"},
		{"github https no suffix", "https://github.com/org/repo", "git@github.com:org/repo.git"},
		{"gitlab https rewrite", "https://gitlab.com/group/sub/repo.git", "git@gitlab.com:group/sub/repo.git"},
		{"bitbucket https rewrite", "https://bitbucket.org/team/repo", "git@bitbucket.org:team/repo.git"},
		{"https with user prefix", "https://user@github.com/org/repo.git", "git@github.com:org/repo.git"},
		{"unknown host left as-is", "https://git.example.com/org/repo.git", "https://git.example.com/org/repo.git"},
		{"trailing slash", "https://github.com/org/repo/", "git@github.com:org/repo.git"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := gitSSHURL(tc.in); got != tc.want {
				t.Errorf("gitSSHURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGitHTTPSURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"https passthrough", "https://github.com/org/repo.git", "https://github.com/org/repo.git"},
		{"https no suffix passthrough", "https://github.com/org/repo", "https://github.com/org/repo"},
		{"github ssh rewrite", "git@github.com:org/repo.git", "https://github.com/org/repo.git"},
		{"github ssh no suffix", "git@github.com:org/repo", "https://github.com/org/repo.git"},
		{"ssh scheme rewrite", "ssh://git@github.com/org/repo.git", "https://github.com/org/repo.git"},
		{"gitlab nested path", "git@gitlab.com:group/sub/repo.git", "https://gitlab.com/group/sub/repo.git"},
		{"bitbucket ssh rewrite", "git@bitbucket.org:team/repo.git", "https://bitbucket.org/team/repo.git"},
		{"unknown host left as-is", "git@git.example.com:org/repo.git", "git@git.example.com:org/repo.git"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := gitHTTPSURL(tc.in); got != tc.want {
				t.Errorf("gitHTTPSURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestChooseCloneURL(t *testing.T) {
	for _, tc := range []struct {
		name       string
		origin     string
		haveSSHKey bool
		wantURL    string
		wantHTTPS  bool
	}{
		{"key, https origin → ssh", "https://github.com/org/repo.git", true, "git@github.com:org/repo.git", false},
		{"key, ssh origin → ssh", "git@github.com:org/repo.git", true, "git@github.com:org/repo.git", false},
		{"no key, https origin → https", "https://github.com/org/repo.git", false, "https://github.com/org/repo.git", true},
		{"no key, ssh origin → https", "git@github.com:org/repo.git", false, "https://github.com/org/repo.git", true},
		{"no key, ssh no suffix → https", "git@github.com:org/repo", false, "https://github.com/org/repo.git", true},
		{"no key, unknown host ssh → unchanged", "git@git.example.com:org/repo.git", false, "git@git.example.com:org/repo.git", false},
		{"no key, unknown host https → unchanged", "https://git.example.com/org/repo.git", false, "https://git.example.com/org/repo.git", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotURL, gotHTTPS := chooseCloneURL(tc.origin, tc.haveSSHKey)
			if gotURL != tc.wantURL || gotHTTPS != tc.wantHTTPS {
				t.Errorf("chooseCloneURL(%q, %v) = (%q, %v), want (%q, %v)",
					tc.origin, tc.haveSSHKey, gotURL, gotHTTPS, tc.wantURL, tc.wantHTTPS)
			}
		})
	}
}

func TestBuildClaudeCommand(t *testing.T) {
	if got, want := buildClaudeCommand("repo", "sid-1"),
		"tmux new-session -A -s claude -c repo 'exec claude --session-id sid-1'"; got != want {
		t.Errorf("got  %q\n want %q", got, want)
	}
	if got, want := buildClaudeCommand("my repo", "sid-1"),
		"tmux new-session -A -s claude -c 'my repo' 'exec claude --session-id sid-1'"; got != want {
		t.Errorf("quoted dir:\n got  %q\n want %q", got, want)
	}
}

func TestBuildResumeCommand(t *testing.T) {
	got := buildResumeCommand("repo", "sid-1")
	// Resumes when the transcript exists, else starts fresh under the same ID
	// (a session created but never talked to has no transcript to resume).
	for _, want := range []string{
		"tmux new-session -A -s claude -c repo ",
		"if ls ~/.claude/projects/*/sid-1.jsonl >/dev/null 2>&1;",
		"then exec claude --resume sid-1;",
		"else exec claude --session-id sid-1;",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("buildResumeCommand missing %q\n got %q", want, got)
		}
	}
}

// The ID comes from a stored record. localsession rejects a non-UUID one at
// load, so this is the second line of defence: it must not be able to close the
// quoting and append a command of its own.
func TestBuildResumeCommandQuotesSessionID(t *testing.T) {
	got := buildResumeCommand("repo", "sid; touch /tmp/pwned; #")
	if strings.Contains(got, "--resume sid; touch /tmp/pwned") {
		t.Errorf("session ID spliced unquoted into the command: %q", got)
	}
	// The leading * must still be left outside the quotes to expand.
	if !strings.Contains(got, "~/.claude/projects/*/'") {
		t.Errorf("quoting swallowed the project glob: %q", got)
	}
}

func TestBuildCloneCommand(t *testing.T) {
	const url = "git@github.com:org/repo.git"
	if got, want := buildCloneCommand(url, "repo", "main", false),
		"GIT_TERMINAL_PROMPT=0 GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=accept-new' git clone --branch main git@github.com:org/repo.git repo"; got != want {
		t.Errorf("with branch:\n got  %q\n want %q", got, want)
	}
	if got, want := buildCloneCommand(url, "repo", "feature/x", true),
		"GIT_TERMINAL_PROMPT=0 GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=accept-new' git clone git@github.com:org/repo.git repo"; got != want {
		t.Errorf("default branch:\n got  %q\n want %q", got, want)
	}
}

func TestBuildGitIdentityCommand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		uName string
		email string
		want  string
	}{
		{"both set", "Ada Lovelace", "ada@example.com",
			"git config --global user.name 'Ada Lovelace' && git config --global user.email ada@example.com"},
		{"name only", "Ada Lovelace", "",
			"git config --global user.name 'Ada Lovelace'"},
		{"email only", "", "ada@example.com",
			"git config --global user.email ada@example.com"},
		{"neither set", "", "", ""},
		{"name with quote is escaped", "O'Brien", "",
			`git config --global user.name 'O'\''Brien'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildGitIdentityCommand(tc.uName, tc.email); got != tc.want {
				t.Errorf("buildGitIdentityCommand(%q, %q):\n got  %q\n want %q", tc.uName, tc.email, got, tc.want)
			}
		})
	}
}

func TestGitIdentityLabel(t *testing.T) {
	for _, tc := range []struct {
		uName string
		email string
		want  string
	}{
		{"Ada Lovelace", "ada@example.com", "Ada Lovelace <ada@example.com>"},
		{"Ada Lovelace", "", "Ada Lovelace"},
		{"", "ada@example.com", "ada@example.com"},
		{"", "", ""},
	} {
		if got := gitIdentityLabel(tc.uName, tc.email); got != tc.want {
			t.Errorf("gitIdentityLabel(%q, %q) = %q, want %q", tc.uName, tc.email, got, tc.want)
		}
	}
}

func TestRepoDirFromURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"ssh form", "git@github.com:org/repo.git", "repo"},
		{"https form", "https://github.com/org/repo.git", "repo"},
		{"no git suffix", "https://github.com/org/repo", "repo"},
		{"nested path", "https://gitlab.com/group/sub/repo.git", "repo"},
		{"trailing slash", "https://github.com/org/repo/", "repo"},
		{"trailing slash after .git", "https://github.com/org/repo.git/", "repo"},
		{"ssh form, trailing slash after .git", "git@github.com:org/repo.git/", "repo"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := repoDirFromURL(tc.in); got != tc.want {
				t.Errorf("repoDirFromURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
