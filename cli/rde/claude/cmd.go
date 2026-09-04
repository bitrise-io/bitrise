// Package claude wires the `bitrise rde claude` command: spin up an
// RDE session and attach straight into Claude Code.
package claude

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v2/internal/app"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
)

// The command provisions a templateless session from a chosen stack + machine
// type, so there's no template to resolve or keep in sync.

// NewCmd returns the `bitrise rde claude` command.
func NewCmd() *cobra.Command {
	var (
		waitTimeout    time.Duration
		resume         bool
		continueLatest bool
		stack          string
		machineType    string
	)

	c := &cobra.Command{
		Use:   "claude [SESSION_ID]",
		Short: "Create an RDE session and attach to Claude Code",
		Long: `Create a fresh RDE session, wait for it to start, then SSH in and drop you
directly into Claude Code (not a shell).

You pick the stack first, then a machine type compatible with it. Your choice
is remembered per repository and preselected next time, so you can just press
Enter. Pass --stack / --machine-type to skip the prompts
(useful for scripts); when stdin isn't a terminal the remembered or default
selection is used without prompting.

Run this from inside a git repository: the session clones the same repository
and branch you're on (via 'git clone') and starts Claude Code inside that
clone. Only the pushed remote state of the branch is cloned — local
uncommitted or unpushed changes are not transferred.

On macOS sessions, while you're in Claude Code it can open a VNC viewer on your
machine showing the session's desktop when you ask to see something visual (a
simulator, app, or browser running in the session). Linux sessions have no
desktop to show, so this isn't offered there; either way the rest of the
command is unaffected.

When you exit Claude Code, the session is terminated automatically (its VM is
torn down), but the session is preserved and can be restored later. Each
invocation creates a new, uniquely named session (claude-<id>).

Resume a previous session instead of creating one:

  --continue        resume the most recent session started from this repo
  --resume          pick a previous session for this repo from a list
  --resume SESSION  resume a specific session by ID (or name)

Resuming reconnects to the session if it's still running, otherwise restores it
and continues the same Claude Code conversation. Sessions are tracked locally
per repository as you use them; while a session is live, its AI-generated title
and a "repo @ branch" description (with the pull-request URL) are kept up to
date both locally and on the session itself.

A local SSH agent ($SSH_AUTH_SOCK), if present, is forwarded into the session
so the clone (and git-over-SSH inside the session) uses your local keys. If the
repo's origin is an HTTPS GitHub/GitLab/Bitbucket URL, it's rewritten to its
SSH form so the forwarded agent can authenticate. Your local git identity
(user.name / user.email) is also copied into the session and set globally, so
commits made there are attributed to you rather than the session's account.

Unless a Claude Code token is already configured on the control plane, a local
credential is saved there before the session is created — taken from
$CLAUDE_CODE_OAUTH_TOKEN or $ANTHROPIC_API_KEY, then ~/.claude/.credentials.json,
or minted with 'claude setup-token' (browser auth). The control plane uses that
token to install Claude Code and tmux during provisioning and to authenticate
the in-session claude; once saved, future sessions reuse it.`,
		Example: `  bitrise rde claude --workspace WORKSPACE_ID
  bitrise rde claude --stack osx-xcode-16.0.x-edge --machine-type g2.mac.m2pro.4c-6g
  bitrise rde claude --continue
  bitrise rde claude --resume`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			// Make the command interrupt-aware: Ctrl-C cancels ctx, which kills
			// any child we launched (e.g. 'claude setup-token') and lets the
			// deferred session cleanup run instead of hard-killing the process.
			// (During the interactive Claude attach the terminal is in raw mode,
			// so Ctrl-C is delivered to the remote claude, not to us — as
			// intended.)
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()

			// A positional argument implies resume-by-ID (fresh runs take no
			// args), so any of --continue, --resume, or an explicit SESSION_ID
			// routes to the resume flow.
			target := ""
			if len(args) == 1 {
				target = args[0]
			}
			if continueLatest || resume || target != "" {
				return runResume(ctx, cmd, resumeOptions{
					target:         target,
					continueLatest: continueLatest,
					waitTimeout:    waitTimeout,
				})
			}
			return runFresh(ctx, cmd, freshOptions{
				waitTimeout: waitTimeout,
				stack:       stack,
				machineType: machineType,
			})
		},
	}

	c.Flags().DurationVar(&waitTimeout, "wait-timeout", 10*time.Minute, "max time to wait for the session to start (uses Go duration syntax: 30s, 5m, 1h)")
	c.Flags().BoolVar(&resume, "resume", false, "resume a previous session for this repo; with no SESSION_ID, pick one from a list")
	c.Flags().BoolVar(&continueLatest, "continue", false, "resume the most recent session started from this repo")
	c.Flags().StringVar(&stack, "stack", "", "stack to use (skips the stack prompt); see 'rde stack list'")
	c.Flags().StringVar(&machineType, "machine-type", "", "machine type to use (skips the machine-type prompt); see 'rde machine-type list'")
	c.MarkFlagsMutuallyExclusive("resume", "continue")
	return c
}

// freshOptions configures a fresh `rde claude` run.
type freshOptions struct {
	waitTimeout time.Duration
	stack       string // --stack override ("" = prompt/default)
	machineType string // --machine-type override ("" = prompt/default)
}

// runFresh creates a new RDE session and attaches to a fresh Claude Code
// session inside it.
func runFresh(ctx context.Context, cmd *cobra.Command, opts freshOptions) error {
	// Resolve the local repo + branch FIRST so a non-repo cwd fails
	// before we create (and would have to tear down) a session.
	detector := internalapp.ExecGitDetector{}
	originURL, err := detector.RemoteURL(ctx)
	if err != nil {
		return fmt.Errorf("detect git remote: %w", err)
	}
	if originURL == "" {
		return fmt.Errorf("current directory is not a git repository (or has no 'origin' remote); rde claude clones the current repo into the session")
	}
	branch, err := detector.CurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("detect git branch: %w", err)
	}
	if branch == "" {
		return fmt.Errorf("could not determine the current git branch (detached HEAD?); check out a branch before running rde claude")
	}
	cloneURL := gitSSHURL(originURL)
	repoDir := repoDirFromURL(originURL)
	repoPath := repoRootPath(ctx)

	workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
	if err != nil {
		return err
	}
	client, err := cmdutil.NewRDEClient(cmd)
	if err != nil {
		return err
	}
	svc := internalrde.NewService(client)

	name, err := generateSessionName()
	if err != nil {
		return err
	}
	// Generate the Claude session ID up front so we can store it, find its
	// transcript to read the AI title, and resume it later.
	claudeSessionID, err := generateClaudeSessionID()
	if err != nil {
		return err
	}

	log := newStepLogger(cmd)

	// ── Stack & machine type ───────────────────────────────────────
	// Choose before auth so all interactive input is gathered up front (and a
	// bad --stack/--machine-type fails fast, before the auth step that can open
	// a browser). The choice is remembered per repo and used on the next run.
	log.group("Stack & machine type")
	stack, stackTitle, machineType, machineLbl, err := selectStackAndMachineType(ctx, cmd, svc, log, workspaceID, repoPath, opts.stack, opts.machineType)
	if errors.Is(err, errSelectionCancelled) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		return nil
	}
	if err != nil {
		return err
	}
	log.done("Stack %s, machine type %s", stackTitle, machineLbl)

	// ── Claude Code auth ───────────────────────────────────────────
	// Ensure a Claude Code token exists on the control plane *before*
	// creating the session: the backend installs claude + tmux during
	// provisioning only when the token saved input is present, and maps
	// it into the session as an env var so claude is authenticated.
	// Without a token the session would be useless for this command, so
	// a failure here aborts before any VM is created.
	log.group("Claude Code auth")
	if err := ensureClaudeAuth(ctx, svc, log); err != nil {
		return err
	}

	// ── Session ────────────────────────────────────────────────────
	log.group("Session")
	log.step("Creating session %q…", name)
	res, err := svc.CreateSession(ctx, workspaceID, internalrde.CreateSessionRequest{
		Name:        name,
		StackID:     stack,
		MachineType: machineType,
		// Auth is guaranteed present by now, so map the token saved
		// input in: provisioning installs claude + tmux and the session
		// is authenticated.
		MapSavedToSessionInputs: true,
	})
	if err != nil {
		return err
	}
	sessionID := res.Session.ID

	// Remember the selection so the next run in this repo preselects it.
	// Best-effort: a failure only costs the preselect convenience.
	if err := localsession.SavePrefs(repoPath, localsession.Prefs{Stack: stack, MachineType: machineType}); err != nil {
		log.warn("Could not save stack/machine-type choice for next time: %v", err)
	}

	// Persist a local record immediately so the session is resumable even if
	// this process dies abruptly (e.g. laptop sleep, SIGKILL) before the
	// metadata monitor enriches it. Best-effort: a failure only costs resume.
	rec := localsession.Record{
		RDESessionID:    sessionID,
		WorkspaceID:     workspaceID,
		Name:            name,
		ClaudeSessionID: claudeSessionID,
		Repo:            originURL,
		RepoPath:        repoPath,
		Branch:          branch,
		RemoteRepoDir:   repoDir,
	}
	if err := localsession.Save(rec); err != nil {
		log.warn("Could not save local session record (resume may be unavailable): %v", err)
	}

	// Terminate the session on every exit path, including PTY errors
	// and Ctrl-C. Terminating stops the VM but preserves the session so
	// it can be restored later. The one exception is leaveRunning: when the
	// user interrupts a reconnect, the VM is left running so they can
	// reconnect later.
	leaveRunning := false
	defer terminateOnExit(svc, log, workspaceID, sessionID, name, &leaveRunning)()

	// The whole startup wait — reaching "running" and then SSH
	// credentials being issued — shares one timeout budget.
	waitCtx, cancel := context.WithTimeout(ctx, opts.waitTimeout)
	defer cancel()

	var ready internalrde.Session
	if err := log.await(waitCtx,
		fmt.Sprintf("Booting session (timeout %s)…", opts.waitTimeout), "Session booted",
		func(c context.Context, status func(string)) error {
			var e error
			ready, e = svc.WaitForReady(c, workspaceID, sessionID, 0, status)
			return e
		}); err != nil {
		return fmt.Errorf("waiting for session: %w", err)
	}
	if ready.Status != "running" {
		cmdutil.SilenceRootErrors(cmd)
		return fmt.Errorf("session ended provisioning with status %q (expected running)", ready.Status)
	}

	// "running" does not mean SSH is reachable yet — the backend
	// issues credentials a few seconds later. Wait for them before
	// dialing in.
	if err := log.await(waitCtx, "Waiting for remote access…", "Remote access ready",
		func(c context.Context, _ func(string)) error {
			_, e := svc.WaitForSSHReady(c, workspaceID, sessionID, 0)
			return e
		}); err != nil {
		return fmt.Errorf("waiting for SSH access: %w", err)
	}
	log.done("Session ready")

	// ── Repository ─────────────────────────────────────────────────
	log.group("Repository")

	// The clone authenticates via the forwarded local SSH agent; make
	// sure an agent is running with a key loaded before we dial in.
	// cleanupAgent kills a temporary agent we may have started; it
	// must outlive the claude session, so defer it to command exit.
	cleanupAgent, repoAuth, haveSSHKey := ensureAgentHasKey(ctx, log, cloneURL)
	defer cleanupAgent()

	// Without a forwarded SSH key an SSH clone fails even for a public repo, so
	// fall back to the repo's HTTPS URL — public repositories clone with no
	// credentials. Private repos still need a key (HTTPS would prompt and fail).
	cloneURL, viaHTTPS := chooseCloneURL(originURL, haveSSHKey)
	if viaHTTPS {
		repoAuth = "HTTPS (no SSH key found; public repos only)"
	}
	log.step("Auth: %s", repoAuth)

	// Clone the same repo + branch into the session over the
	// forwarded SSH agent. Runs in its own interactive session so its
	// output (git progress) streams live with no timeout.
	// StrictHostKeyChecking=accept-new avoids a host-key prompt that
	// would otherwise hang the non-interactive clone.
	//
	// If the current branch hasn't been pushed (not on the remote),
	// fall back to cloning the remote's default branch.
	useDefaultBranch := false
	if found, determined := remoteHasBranch(ctx, branch); determined && !found {
		useDefaultBranch = true
	}
	if useDefaultBranch {
		log.warn("Branch %q is not on the remote; cloning the default branch instead.", branch)
		log.step("Cloning %s (default branch)…", repoDir)
	} else {
		log.step("Cloning %s (branch %s)…", repoDir, branch)
	}
	cloneCmd := buildCloneCommand(cloneURL, repoDir, branch, useDefaultBranch)
	// Indent the clone's streamed output so it lines up under the
	// group. (The interactive Claude attach below streams raw — a
	// full-screen TUI can't be column-shifted.)
	cloneCode, err := svc.ExecuteInteractive(ctx, workspaceID, sessionID, cloneCmd, os.Stdin, newIndentWriter(os.Stdout), newIndentWriter(os.Stderr))
	if err != nil {
		return err
	}
	if cloneCode != 0 {
		cmdutil.SilenceRootErrors(cmd)
		return fmt.Errorf("git clone failed (status %d)", cloneCode)
	}

	// Mirror the user's local git identity onto the session so commits Claude
	// makes are attributed to them, not the session's system account. Part of
	// the Repository group, best-effort (see syncGitIdentity).
	syncGitIdentity(ctx, svc, log, workspaceID, sessionID)
	log.done("Repository cloned")

	// ── Claude Code ────────────────────────────────────────────────
	// Auth was provisioned up front (token saved input → session env
	// var), so just start claude — no credential injection here.
	log.group("Claude Code")
	log.step("Starting…")
	// Start claude inside a tmux session (in the cloned repo) so it
	// survives a disconnect and can be reattached later; the user still
	// lands directly in claude, not a shell.
	exitCode, err := attachClaude(ctx, svc, log, attachParams{
		workspaceID:     workspaceID,
		sessionID:       sessionID,
		claudeSessionID: claudeSessionID,
		claudeCmd:       buildClaudeCommand(repoDir, claudeSessionID),
		record:          rec,
		describe:        newDescriber(repoSlugFromURL(originURL), branch),
	})
	if errors.Is(err, errReconnectInterrupted) {
		leaveRunning = true // skip termination; deferred cleanup leaves the VM up
		return nil
	}
	if err != nil {
		return err
	}
	return claudeExitError(cmd, exitCode)
}

// attachParams bundles what attachClaude needs to run the interactive Claude
// session and keep the session's metadata up to date.
type attachParams struct {
	workspaceID     string
	sessionID       string
	claudeSessionID string
	claudeCmd       string
	record          localsession.Record
	describe        func(context.Context) string
}

// attachClaude starts the metadata monitor for the session and runs the
// interactive Claude command, reattaching on a dropped connection. It returns
// Claude's exit code. The monitor is cancelled when this returns.
func attachClaude(ctx context.Context, svc *internalrde.Service, log *stepLogger, p attachParams) (int, error) {
	// The monitor reads the AI title from the session and pushes name +
	// description updates to the local store and the API. Tie it to a child
	// context so it stops as soon as the interactive session ends.
	monCtx, monCancel := context.WithCancel(ctx)
	defer monCancel()
	mon := &internalrde.ClaudeMetadataMonitor{
		Service:         svc,
		WorkspaceID:     p.workspaceID,
		SessionID:       p.sessionID,
		ClaudeSessionID: p.claudeSessionID,
		Interval:        internalrde.DefaultMetadataInterval,
		Record:          p.record,
		Describe:        p.describe,
	}
	go mon.Run(monCtx)

	// Best-effort host bridge: lets the in-session Claude trigger local actions
	// (transfer files to/from the user's machine, and on VNC-capable sessions
	// open a viewer showing the session's desktop). The action set depends on the
	// session — transfers apply everywhere, open-vnc only to sessions that expose
	// a VNC endpoint (macOS today; Linux sessions have none). Start it — and write
	// its skill — before the interactive attach, so the capability is in place
	// when Claude launches; it degrades silently if the session does not permit
	// the reverse forward, and never disrupts the attach. localDir is the local
	// working dir, used to resolve relative transfer paths.
	localDir := p.record.RepoPath
	if localDir == "" {
		if wd, err := os.Getwd(); err == nil {
			localDir = wd
		}
	}
	if actions := localHostActions(monCtx, svc, p.workspaceID, p.sessionID, localDir); len(actions) > 0 {
		bridge := newHostBridge(svc, p.workspaceID, p.sessionID, actions)
		if bridgeErr := bridge.Start(monCtx); bridgeErr != nil {
			log.step("Host actions unavailable in this session")
		} else {
			defer bridge.Close()
			go bridge.Serve(monCtx)
			log.step("%s", hostActionsMessage(actions))
		}
	}

	exitCode, err := svc.ExecuteInteractive(ctx, p.workspaceID, p.sessionID, p.claudeCmd, os.Stdin, os.Stdout, os.Stderr)
	if errors.Is(err, internalrde.ErrConnectionLost) {
		// The connection dropped but claude keeps running in tmux on the
		// session — reattach instead of tearing it down.
		log.group("Reconnecting")
		log.warn("Connection lost; Claude Code is still running in tmux on the session.")
		exitCode, err = reattachWithRetry(ctx, svc, log, p.workspaceID, p.sessionID)
	}
	return exitCode, err
}

// claudeExitError maps Claude's exit code to a user-facing error (nil on 0).
func claudeExitError(cmd *cobra.Command, exitCode int) error {
	if exitCode == 0 {
		return nil
	}
	cmdutil.SilenceRootErrors(cmd)
	if exitCode == 127 {
		// 127 = "command not found" from the login shell.
		return fmt.Errorf("could not start Claude Code: 'tmux' or 'claude' is not installed on the session")
	}
	return fmt.Errorf("claude exited with status %d", exitCode)
}

// terminateOnExit returns a cleanup that terminates the session. Terminating
// stops the VM but preserves the session so it can be restored later. A fresh
// context is used because the command context may already be cancelled (e.g.
// Ctrl-C) by the time cleanup runs.
//
// When *leaveRunning is true the VM is left running instead — the user
// interrupted a reconnect (see errReconnectInterrupted), so claude keeps
// running in tmux and they can reconnect later. leaveRunning may be nil, which
// reads as false (always terminate).
func terminateOnExit(svc *internalrde.Service, log *stepLogger, workspaceID, sessionID, name string, leaveRunning *bool) func() {
	return func() {
		if leaveRunning != nil && *leaveRunning {
			log.group("Disconnected")
			log.step("Left session %q running; reconnect with 'rde claude --resume %s'.", name, sessionID)
			return
		}
		log.group("Cleanup")
		log.step("Terminating session %q…", name)
		termCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := svc.TerminateSession(termCtx, workspaceID, sessionID); err != nil {
			log.warn("Failed to terminate session %q: %v", sessionID, err)
			return
		}
		log.done("Session terminated")
	}
}

// generateSessionName returns a unique "claude-<hex>" session name. There is
// no UUID dependency in this module, so a short random suffix from crypto/rand
// is used to keep names collision-free across invocations.
func generateSessionName() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session name: %w", err)
	}
	return "claude-" + hex.EncodeToString(b), nil
}

// remoteHasBranch reports whether branch exists on the cwd repo's "origin"
// remote. `git ls-remote --exit-code` exits 0 when the ref is found and 2 when
// it isn't; any other outcome (network/auth error, git missing) leaves it
// undetermined, in which case the caller keeps the requested branch.
func remoteHasBranch(ctx context.Context, branch string) (found, determined bool) {
	// branch comes from the local repo's checked-out HEAD, passed as its own argv element — no shell, no injection
	err := exec.CommandContext(ctx, "git", "ls-remote", "--heads", "--exit-code", "origin", branch).Run()
	if err == nil {
		return true, true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
		return false, true
	}
	return false, false
}

// buildCloneCommand returns the remote shell command that clones cloneURL into
// repoDir. When useDefaultBranch is true it omits --branch, so git clones the
// remote's default branch (the fallback when the local branch isn't pushed).
// GIT_SSH_COMMAND's accept-new keeps the non-interactive clone from hanging on
// a host-key prompt; GIT_TERMINAL_PROMPT=0 makes an HTTPS clone of a private
// repo (no credentials) fail fast instead of hanging on a username prompt.
func buildCloneCommand(cloneURL, repoDir, branch string, useDefaultBranch bool) string {
	env := "GIT_TERMINAL_PROMPT=0 GIT_SSH_COMMAND=" + cmdutil.ShellQuote("ssh -o StrictHostKeyChecking=accept-new")
	if useDefaultBranch {
		return fmt.Sprintf("%s git clone %s %s",
			env, cmdutil.ShellQuote(cloneURL), cmdutil.ShellQuote(repoDir))
	}
	return fmt.Sprintf("%s git clone --branch %s %s %s",
		env, cmdutil.ShellQuote(branch), cmdutil.ShellQuote(cloneURL), cmdutil.ShellQuote(repoDir))
}

// syncGitIdentity copies the user's local git identity (user.name/user.email)
// onto the session, set globally (~/.gitconfig), so commits Claude makes are
// attributed to the user rather than the session's system account
// (vagrant@…/ubuntu@… — the SSH users sessions run as). It complements the
// forwarded SSH agent: the agent lets the session push, this makes the commits
// it pushes carry the right author.
//
// Best-effort by design: when no local identity is set it does nothing, and a
// remote failure only warns — neither blocks the session. Idempotent and
// global, so resuming re-applies it (covering an identity that changed locally,
// or sessions created before this existed) and it survives a restore since ~ is
// on the persistent disk.
func syncGitIdentity(ctx context.Context, svc *internalrde.Service, log *stepLogger, workspaceID, sessionID string) {
	name, email := localGitIdentity(ctx)
	command := buildGitIdentityCommand(name, email)
	if command == "" {
		return // no local identity configured — leave the session's default alone
	}
	log.step("Setting git identity (%s)…", gitIdentityLabel(name, email))
	res, err := svc.Execute(ctx, workspaceID, sessionID, command, nil, internalrde.DefaultExecuteTimeout)
	switch {
	case err != nil:
		log.warn("Could not set git identity on the session: %v", err)
	case res.ExitCode != 0:
		log.warn("Could not set git identity on the session (status %d): %s", res.ExitCode, strings.TrimSpace(res.Stderr))
	}
}

// localGitIdentity reads the user's git identity (user.name and user.email)
// from the cwd repo. Reading from inside the repo honors git's full precedence
// (repo-local > global > system), so it matches who commits are attributed to
// here. Either value comes back "" when it isn't set.
func localGitIdentity(ctx context.Context) (name, email string) {
	return gitConfigValue(ctx, "user.name"), gitConfigValue(ctx, "user.email")
}

// gitConfigValue returns the configured value for a git config key, or "" if
// it isn't set (or git errors).
func gitConfigValue(ctx context.Context, key string) string {
	// key is a hardcoded git config name, never user input
	out, err := exec.CommandContext(ctx, "git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// buildGitIdentityCommand returns the remote shell command that sets the user's
// git identity globally on the session, or "" when neither value is set. Each
// value is shell-quoted — names legitimately contain spaces, and emails are
// taken from local config unvalidated.
func buildGitIdentityCommand(name, email string) string {
	var parts []string
	if name != "" {
		parts = append(parts, "git config --global user.name "+cmdutil.ShellQuote(name))
	}
	if email != "" {
		parts = append(parts, "git config --global user.email "+cmdutil.ShellQuote(email))
	}
	return strings.Join(parts, " && ")
}

// gitIdentityLabel renders an identity for a progress line: "Name <email>", or
// just whichever half is set.
func gitIdentityLabel(name, email string) string {
	switch {
	case name != "" && email != "":
		return name + " <" + email + ">"
	case email != "":
		return email
	default:
		return name
	}
}

// buildClaudeCommand returns the remote shell command that starts Claude Code
// inside repoDir, running it in a tmux session so it survives an SSH
// disconnect and can be reattached later. Auth comes from the session's
// environment (the control-plane token saved input mapped in as an env var),
// so nothing is injected here.
//
//   - `tmux new-session -A -s claude`: attach to the "claude" session if it
//     already exists, else create it.
//   - `-c <repoDir>`: the pane starts in the cloned repo.
//   - `exec claude --session-id <id>`: the pane's shell becomes claude (so the
//     tmux session ends when claude exits), started with a known session ID so
//     we can locate its transcript and resume it later.
func buildClaudeCommand(repoDir, claudeSessionID string) string {
	return buildTmuxClaudeCommand(repoDir, "exec claude --session-id "+claudeSessionID)
}

// buildResumeCommand is buildClaudeCommand's resume counterpart: it resumes the
// existing Claude conversation (`claude --resume <id>`) in repoDir. The `-A`
// flag means that if the "claude" tmux session is still alive (the VM was never
// torn down — e.g. after a laptop sleep), tmux reattaches to the live claude
// and the command is ignored; only on a fresh VM (after a restore) does the
// inner command run.
//
// Claude Code only persists a session once its conversation starts, so a
// session that was created but never talked to has no transcript and
// `claude --resume <id>` fails ("No conversation found"). The inner command
// guards against that: it resumes only when the transcript exists, otherwise it
// starts fresh under the SAME session ID — so the metadata monitor and any
// later resume keep working. The transcript glob matches the file claude writes
// on the first message (the same path readAITitleCommand reads).
func buildResumeCommand(repoDir, claudeSessionID string) string {
	id := claudeSessionID
	inner := fmt.Sprintf(
		"if ls ~/.claude/projects/*/%s.jsonl >/dev/null 2>&1; "+
			"then exec claude --resume %s; "+
			"else exec claude --session-id %s; fi",
		id, id, id)
	return buildTmuxClaudeCommand(repoDir, inner)
}

// buildTmuxClaudeCommand wraps a claude invocation in the shared tmux launcher.
// inner must be a single shell command (run via the shell so `exec` works).
func buildTmuxClaudeCommand(repoDir, inner string) string {
	return "tmux new-session -A -s claude -c " + cmdutil.ShellQuote(repoDir) + " " + cmdutil.ShellQuote(inner)
}

// buildReattachCommand reattaches to the running "claude" tmux session after a
// dropped connection. It does NOT recreate the session: if claude already
// exited, tmux exits non-zero and we stop rather than starting claude over.
func buildReattachCommand() string {
	return "tmux attach-session -t claude"
}

// ensureClaudeAuth makes sure a Claude Code token is configured on the control
// plane (a CLAUDE_CODE_OAUTH_TOKEN / ANTHROPIC_API_KEY saved input). When it's
// missing, it resolves a local credential — env var, ~/.claude/.credentials.json,
// or a freshly minted `claude setup-token` (browser auth) — and saves it.
//
// A token is mandatory: it's what makes the backend install claude + tmux and
// authenticate the in-session claude, so the session is useless without one.
// Any failure (control plane unreachable, no credential found, setup-token
// cancelled/declined, save failed, or the run interrupted) returns an error so
// the caller aborts before creating a VM.
func ensureClaudeAuth(ctx context.Context, svc *internalrde.Service, log *stepLogger) error {
	has, err := controlPlaneHasClaudeToken(ctx, svc)
	if err != nil {
		// We can't tell whether a token is configured. The same API is needed
		// to save the token and create the session, so abort with a clear
		// message rather than detouring through an unexpected setup-token flow.
		return fmt.Errorf("check Claude Code token on the control plane: %w", err)
	}
	if has {
		log.step("Using the Claude Code token already on the control plane")
		return nil
	}

	cred, ok := existingLocalCredential()
	if !ok {
		log.step("No token on the control plane or locally; minting one with 'claude setup-token'…")
		cred, ok = mintSetupToken(ctx)
	}
	if !ok {
		// A cancelled/interrupted run surfaces as the context error so the
		// user sees "interrupted" rather than a misleading "no token" message.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("no Claude Code token available ('claude setup-token' did not return one); not creating a session. " +
			"Set $CLAUDE_CODE_OAUTH_TOKEN or $ANTHROPIC_API_KEY, log in locally, or configure a token on the control plane, then retry")
	}

	if _, err := svc.CreateSavedInput(ctx, internalrde.CreateSavedInputRequest{
		Key: cred.EnvVar, Value: cred.Value, IsSecret: true,
	}); err != nil {
		return fmt.Errorf("save Claude Code token on the control plane: %w", err)
	}
	log.step("Saved %s on the control plane (from %s)", cred.EnvVar, cred.Source)
	return nil
}

const reconnectInterval = 3 * time.Second

// errReconnectInterrupted signals that the user pressed Ctrl-C while we were
// retrying to reconnect after a dropped connection. Unlike every other exit
// path, this one deliberately LEAVES the session's VM running so the user can
// reconnect later (e.g. 'rde claude --continue'); the command treats it as a
// clean, non-error exit. (Ctrl-C only reaches us between/around reconnect
// attempts — while attached, the terminal is raw and Ctrl-C goes to claude.)
var errReconnectInterrupted = errors.New("reconnect interrupted")

// reattachWithRetry keeps reattaching to the session's tmux after the
// connection drops, retrying indefinitely until claude exits, a fatal error
// occurs, or the user interrupts with Ctrl-C. Each attempt's dial is itself
// retried inside ExecuteInteractive, so a brief outage resolves on the next
// attempt. On interrupt it returns errReconnectInterrupted so the caller leaves
// the session running rather than tearing it down.
func reattachWithRetry(ctx context.Context, svc *internalrde.Service, log *stepLogger, workspaceID, sessionID string) (int, error) {
	for attempt := 1; ; attempt++ {
		if ctx.Err() != nil {
			return -1, errReconnectInterrupted
		}
		log.step("Reattaching to Claude Code (attempt %d)…", attempt)
		code, err := svc.ExecuteInteractive(ctx, workspaceID, sessionID, buildReattachCommand(), os.Stdin, os.Stdout, os.Stderr)
		if !errors.Is(err, internalrde.ErrConnectionLost) {
			// A cancelled dial (Ctrl-C mid-attempt) surfaces as a non-ErrConnectionLost
			// error wrapping ctx.Err(); treat it as an interrupt, not a fatal error.
			if ctx.Err() != nil {
				return -1, errReconnectInterrupted
			}
			return code, err // clean exit, real exit code, or a fatal error
		}
		// Still unreachable — wait and try again, unless interrupted meanwhile.
		select {
		case <-ctx.Done():
			return -1, errReconnectInterrupted
		case <-time.After(reconnectInterval):
		}
	}
}

// sshRewriteHosts are the hosts whose HTTPS clone URLs we rewrite to SSH form,
// so the forwarded SSH agent can authenticate the in-session clone. Other
// hosts are left untouched (their HTTPS URLs only work for public repos).
var sshRewriteHosts = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"bitbucket.org": true,
}

// gitSSHURL rewrites an HTTPS clone URL for a known host into its SSH form
// (https://github.com/org/repo(.git) → git@github.com:org/repo.git) so the
// forwarded agent's keys authenticate. URLs that are already SSH (git@… or
// ssh://…), or HTTPS for an unknown host, are returned unchanged.
func gitSSHURL(raw string) string {
	if strings.HasPrefix(raw, "git@") || strings.HasPrefix(raw, "ssh://") {
		return raw
	}
	const httpsPrefix = "https://"
	if !strings.HasPrefix(raw, httpsPrefix) {
		return raw
	}
	rest := strings.TrimPrefix(raw, httpsPrefix)
	host, path, ok := strings.Cut(rest, "/")
	if !ok || path == "" {
		return raw
	}
	// Strip any "user@" credential prefix on the host (e.g. user@github.com).
	if _, h, found := strings.Cut(host, "@"); found {
		host = h
	}
	if !sshRewriteHosts[host] {
		return raw
	}
	path = strings.TrimSuffix(path, "/")
	if !strings.HasSuffix(path, ".git") {
		path += ".git"
	}
	return "git@" + host + ":" + path
}

// gitHTTPSURL rewrites a known-host SSH clone URL into its HTTPS form
// (git@github.com:org/repo.git → https://github.com/org/repo.git) so a public
// repository clones without an SSH key. URLs that are already HTTPS, or that
// target a host we don't recognise, are returned unchanged.
func gitHTTPSURL(raw string) string {
	if strings.HasPrefix(raw, "https://") {
		return raw
	}
	host := sshHostFromURL(raw)
	if !sshRewriteHosts[host] {
		return raw
	}
	slug := repoSlugFromURL(raw)
	if slug == "" {
		return raw
	}
	return "https://" + host + "/" + slug + ".git"
}

// chooseCloneURL picks the URL the in-session clone should use. With a forwarded
// SSH key it prefers the SSH form (so private repos clone); without one it falls
// back to the repo's HTTPS form so public repositories still clone. Unknown
// hosts (no SSH⇄HTTPS rewrite) are left as-is. viaHTTPS reports the fallback.
func chooseCloneURL(originURL string, haveSSHKey bool) (cloneURL string, viaHTTPS bool) {
	ssh := gitSSHURL(originURL)
	if haveSSHKey || !isSSHCloneURL(ssh) {
		return ssh, false
	}
	if https := gitHTTPSURL(originURL); strings.HasPrefix(https, "https://") {
		return https, true
	}
	return ssh, false
}

// repoDirFromURL derives the directory name `git clone` would create from a
// clone URL: the last path segment with any ".git" suffix removed (e.g.
// git@github.com:org/repo.git → repo, https://host/org/repo → repo).
func repoDirFromURL(raw string) string {
	s := strings.TrimSuffix(raw, ".git")
	s = strings.TrimRight(s, "/")
	if i := strings.LastIndexAny(s, "/:"); i >= 0 {
		s = s[i+1:]
	}
	return s
}
