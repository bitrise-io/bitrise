package claude

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// ensureAgentHasKey is an rde-claude-only convenience. The in-session git
// clone authenticates via the forwarded local SSH agent (see exec_ssh.go), so
// an agent with no keys loaded means a private clone fails. It:
//   - starts a temporary agent if none is running (and returns a cleanup that
//     kills it on exit);
//   - adds a key when the agent is empty — chosen from the SSH config for the
//     clone host (via `ssh -G`, which also yields the built-in defaults),
//     falling back to the default key files.
//
// Entirely best-effort: any failure just means the clone may prompt or fail,
// which the user sees live. The returned cleanup is never nil; auth is a short
// description of the resulting git-clone auth method, for display; haveKey
// reports whether the forwarded agent ends up with a usable key (the caller
// falls back to HTTPS for the clone when it doesn't).
func ensureAgentHasKey(ctx context.Context, log *stepLogger, cloneURL string) (cleanup func(), auth string, haveKey bool) {
	cleanup = func() {}

	// Agent forwarding only helps SSH remotes; HTTPS clones don't use it.
	if !isSSHCloneURL(cloneURL) {
		return cleanup, "HTTPS remote (no SSH key forwarding)", false
	}
	// ssh-add may prompt for a passphrase, so only attempt it with a real
	// terminal (the rest of rde claude needs one anyway).
	if !cmdutil.IsTerminal(os.Stdin) {
		desc, have := forwardedAgentDesc(ctx)
		return cleanup, desc, have
	}

	switch agentState(ctx) {
	case agentHasKeys, agentUnknown:
		desc, have := forwardedAgentDesc(ctx)
		return cleanup, desc, have
	case agentNoAgent:
		c, err := startAgent(ctx)
		if err != nil {
			log.warn("No SSH agent running and could not start one (%v).", err)
			desc, have := forwardedAgentDesc(ctx)
			return cleanup, desc, have
		}
		log.step("Started a temporary SSH agent")
		cleanup = c
		// fall through to add a key to the fresh agent
	case agentEmpty:
		// fall through to add a key
	}

	key := pickIdentityFile(ctx, sshHostFromURL(cloneURL))
	if key == "" {
		log.warn("SSH agent has no keys and no key file was found to add.")
		desc, have := forwardedAgentDesc(ctx)
		return cleanup, desc, have
	}
	log.step("Adding SSH key %s to the agent…", key)
	// Indent ssh-add's output (e.g. "Identity added: …") under the group, like
	// the clone. ssh-add writes prompts/results to stderr, so point stdout
	// there too; using one shared writer makes exec serialize writes to it (no
	// concurrent-write race on the indenter).
	out := newIndentWriter(os.Stderr)
	// key is a local key-file path (from ssh -G or default files), passed as its own argv element — no shell, no injection
	add := exec.CommandContext(ctx, "ssh-add", key)
	add.Stdin = os.Stdin
	add.Stdout = out
	add.Stderr = out
	_ = add.Run() // best-effort; the clone surfaces any remaining auth error
	desc, have := forwardedAgentDesc(ctx)
	return cleanup, desc, have
}

// forwardedAgentDesc describes the git-clone auth method based on what the
// local SSH agent currently holds (and will therefore forward into the VM), and
// reports whether it holds at least one key.
func forwardedAgentDesc(ctx context.Context) (desc string, haveKey bool) {
	switch n := agentKeyCount(ctx); {
	case n == 1:
		return "forwarded SSH agent (1 key)", true
	case n > 1:
		return fmt.Sprintf("forwarded SSH agent (%d keys)", n), true
	default:
		return "no forwarded SSH agent keys", false
	}
}

// agentKeyCount returns how many identities the local SSH agent holds, or 0
// when no agent is reachable / it's empty (`ssh-add -l` exits non-zero).
func agentKeyCount(ctx context.Context) int {
	out, err := exec.CommandContext(ctx, "ssh-add", "-l").Output()
	if err != nil {
		return 0
	}
	n := 0
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

// startAgent launches a temporary `ssh-agent`, exports its socket (and PID)
// into this process's environment so ssh-add and our own agent forwarding
// (dialLocalAgent reads $SSH_AUTH_SOCK) use it, and returns a cleanup that
// terminates the agent. The agent must outlive both the clone and the
// interactive claude session, so the caller defers cleanup to process exit.
func startAgent(ctx context.Context) (func(), error) {
	// `ssh-agent -s` forks a daemon and exits, so from here on every error path
	// has an agent to clean up — killAgent, not a bare return.
	out, err := exec.CommandContext(ctx, "ssh-agent", "-s").Output()
	if err != nil {
		return nil, err
	}
	sock := parseAgentVar(string(out), "SSH_AUTH_SOCK")
	pid := parseAgentVar(string(out), "SSH_AGENT_PID")
	if sock == "" {
		killAgent(pid, sock)
		return nil, errors.New("ssh-agent did not report a socket path")
	}
	if err := os.Setenv("SSH_AUTH_SOCK", sock); err != nil {
		killAgent(pid, sock)
		return nil, err
	}
	if pid != "" {
		_ = os.Setenv("SSH_AGENT_PID", pid)
	}
	return func() { killAgent(pid, sock) }, nil
}

// killAgent terminates the agent we started and removes its socket. The env is
// passed explicitly rather than inherited: `ssh-agent -k` reads SSH_AGENT_PID/
// SSH_AUTH_SOCK, and on the failure paths above we have not set them on the
// process yet. Not ctx-bound — ctx is likely already cancelled by the time the
// success path's cleanup runs.
func killAgent(pid, sock string) {
	if pid == "" && sock == "" {
		return
	}
	c := exec.Command("ssh-agent", "-k")
	c.Env = append(os.Environ(), "SSH_AGENT_PID="+pid, "SSH_AUTH_SOCK="+sock)
	_ = c.Run()
}

// parseAgentVar extracts NAME's value from `ssh-agent -s` output, whose lines
// look like `SSH_AUTH_SOCK=/tmp/…/agent.123; export SSH_AUTH_SOCK;`.
func parseAgentVar(out, name string) string {
	for line := range strings.SplitSeq(out, "\n") {
		for field := range strings.SplitSeq(line, ";") {
			if rest, ok := strings.CutPrefix(strings.TrimSpace(field), name+"="); ok {
				return rest
			}
		}
	}
	return ""
}

type agentStatus int

const (
	agentHasKeys agentStatus = iota
	agentEmpty
	agentNoAgent
	agentUnknown
)

// agentState classifies `ssh-add -l`: exit 0 = has identities, exit 1 = agent
// reachable but empty, exit 2 = no agent reachable.
func agentState(ctx context.Context) agentStatus {
	err := exec.CommandContext(ctx, "ssh-add", "-l").Run()
	if err == nil {
		return agentHasKeys
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		switch exitErr.ExitCode() {
		case 1:
			return agentEmpty
		case 2:
			return agentNoAgent
		}
	}
	return agentUnknown
}

// pickIdentityFile returns the first existing private key ssh would use for
// host (from `ssh -G`, which lists configured IdentityFile entries followed by
// the built-in defaults), falling back to the default key files directly when
// `ssh -G` is unavailable.
func pickIdentityFile(ctx context.Context, host string) string {
	var candidates []string
	if host != "" {
		candidates = append(candidates, sshConfigIdentityFiles(ctx, host)...)
	}
	candidates = append(candidates, defaultKeyFiles()...)
	for _, c := range candidates {
		if c != "" && fileExists(c) {
			return c
		}
	}
	return ""
}

// sshConfigIdentityFiles asks `ssh -G git@host` for the identity files ssh
// would try for that host, in order. Paths are returned with ~ expanded.
func sshConfigIdentityFiles(ctx context.Context, host string) []string {
	// host comes from the repo's git remote, passed as its own argv element — no shell, no injection
	out, err := exec.CommandContext(ctx, "ssh", "-G", "git@"+host).Output()
	if err != nil {
		return nil
	}
	var files []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if rest, ok := strings.CutPrefix(line, "identityfile "); ok {
			files = append(files, expandHome(strings.TrimSpace(rest)))
		}
	}
	return files
}

func defaultKeyFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	names := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
	files := make([]string, 0, len(names))
	for _, n := range names {
		files = append(files, filepath.Join(home, ".ssh", n))
	}
	return files
}

// isSSHCloneURL reports whether u is an SSH-form git URL (git@host:… or
// ssh://…), the only form the forwarded agent can authenticate.
func isSSHCloneURL(u string) bool {
	return strings.HasPrefix(u, "git@") || strings.HasPrefix(u, "ssh://")
}

// sshHostFromURL extracts the host from a git clone URL (ssh or https form).
func sshHostFromURL(u string) string {
	s := u
	for _, p := range []string{"ssh://", "https://", "http://", "git://"} {
		s = strings.TrimPrefix(s, p)
	}
	if _, after, ok := strings.Cut(s, "@"); ok {
		s = after
	}
	if i := strings.IndexAny(s, ":/"); i >= 0 {
		s = s[:i]
	}
	return s
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(p, "~"), "/"))
		}
	}
	return p
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
