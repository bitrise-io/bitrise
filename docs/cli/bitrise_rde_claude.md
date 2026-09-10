## bitrise rde claude

Create an RDE session and attach to Claude Code

### Synopsis

Create a fresh RDE session, wait for it to start, then SSH in and drop you
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

  --continue         resume the most recent session started from this repo
  --resume           pick a previous session for this repo from a list
  SESSION            resume a specific session by ID (or name), passed as a
                     plain positional argument, e.g. 'bitrise rde claude SESSION_ID'
                     — not a value for --resume, which takes none

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
the in-session claude; once saved, future sessions reuse it.

```
bitrise rde claude [SESSION_ID] [flags]
```

### Examples

```
  bitrise rde claude --workspace WORKSPACE_ID
  bitrise rde claude --stack osx-xcode-16.0.x-edge --machine-type g2.mac.m2pro.4c-6g
  bitrise rde claude --continue
  bitrise rde claude --resume
  bitrise rde claude SESSION_ID
```

### Options

```
      --continue                resume the most recent session started from this repo
  -h, --help                    help for claude
      --machine-type string     machine type to use (skips the machine-type prompt); see 'rde machine-type list'
      --resume                  resume a previous session for this repo; with no SESSION_ID, pick one from a list
      --stack string            stack to use (skips the stack prompt); see 'rde stack list'
      --wait-timeout duration   max time to wait for the session to start (uses Go duration syntax: 30s, 5m, 1h) (default 10m0s)
      --workspace string        workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
```

### Options inherited from parent commands

```
      --ci              If true it indicates that we're used by another tool so don't require any user input!
      --debug           If true it enables DEBUG mode.
      --no-color        Disable ANSI colors (the NO_COLOR env var is also honored).
  -o, --output string   Output format for commands that support it. Accepted: raw (default), json, yml (alias "human").
      --pr              If true bitrise runs in pull request mode.
  -q, --quiet           Suppress non-error diagnostic messages.
      --theme string    Color theme. Accepted: auto, dark, light, none.
```

### SEE ALSO

* [bitrise rde](bitrise_rde.md)	 - Manage Bitrise Remote Dev Environments (sessions, templates, …)

