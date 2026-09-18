## bitrise rde session exec

Run a command on a session over SSH

### Synopsis

Run a command on a session over SSH and capture its output.

The command runs in a forced-interactive login bash shell (bash -i -l -c) so
the session's PATH, brew-installed binaries, git-lfs, and language version
managers (nvm, pyenv, rbenv, asdf) are all loaded.

By default the tokens after '--' are treated as a program plus literal
arguments: each is passed through verbatim, so shell metacharacters (;, &&,
|, $(...), redirection) are NOT interpreted — 'exec ID -- echo "a; b"' runs
echo with the single literal argument 'a; b'. This keeps quoted arguments
intact (e.g. -m "a message"). (Quote such arguments so your local shell
hands them over as one token rather than splitting them itself.)

Pass --shell to interpret everything after '--' as a shell command line
instead, so pipes, &&, command substitution and redirection work:

  bitrise rde session exec SESSION_ID --shell -- 'cd repo && xcodebuild | xcpretty'

This is the first-class replacement for hand-wrapping every command in
bash -lc "…". Quote the command (or its metacharacters) so your local shell
passes them through unchanged. --shell must come before '--'; after '--' it
is just another literal token.

If a local SSH agent is available ($SSH_AUTH_SOCK set), it's forwarded into
the session — git-over-SSH inside the session uses the caller's local keys.

Local environment variables can be forwarded to the remote command with
--env (repeatable, before '--'): --env NAME forwards the local value of
NAME (an error if unset), --env NAME=VALUE sets a literal. A repo can pin
a shared list in .bitrise/rde.yml (found in the working directory or any
ancestor):

  exec:
    env:
      - API_BASE_URL
      - NPM_TOKEN=abc123

File entries use the same NAME / NAME=VALUE forms, except a NAME that is
unset locally is skipped with a warning instead of failing, so a shared
file never breaks a teammate. --env overrides a same-named file entry;
--no-env-file skips the file entirely. Commit the file to share the list
with your team, or gitignore it to keep a personal list (including
NAME=VALUE literals) out of the repo. Forwarded variables are announced
by name on stderr (silence with -q) — values are never printed locally,
but they do ride in the remote command line, so they are visible in ps
on the session while the command runs.

In raw mode: stdout streams to this CLI's stdout, stderr to stderr, and
this CLI exits non-zero when the remote command exits non-zero.

In --format json/yml mode: a single exit_code/stdout/stderr object is
emitted to stdout regardless of the command's exit status.

The remote command is capped at 10 minutes by default; raise it with --timeout
(e.g. --timeout 20m for a cold xcodebuild) or pass --timeout 0 to disable the
cap. exec holds the SSH connection open for the whole run, so the command dies
if the connection drops — for fire-and-forget work that must outlive the
connection, nohup it inside the session instead.

```
bitrise rde session exec SESSION_ID -- COMMAND [ARGS...] [flags]
```

### Examples

```
  bitrise rde session exec SESSION_ID -- echo hello
  bitrise rde session exec SESSION_ID -- npm test
  bitrise rde session exec SESSION_ID -- git commit -m "a message"
  bitrise rde session exec SESSION_ID --shell -- 'cd repo && ls | head'
  bitrise rde session exec SESSION_ID --env NPM_TOKEN --env CI=1 -- npm test
  bitrise rde session exec SESSION_ID --timeout 20m -- ./scripts/cold-build.sh
  bitrise rde session exec SESSION_ID --format json -- ls -la /opt
```

### Options

```
      --env stringArray    environment variable for the remote command: NAME forwards the local value (errors if unset), NAME=VALUE sets a literal (repeatable; must come before '--')
  -f, --format string      Output format. Accepted: raw (default), json, yml
  -h, --help               help for exec
      --no-env-file        skip reading forwarded env vars from .bitrise/rde.yml
      --shell              interpret everything after '--' as a shell command line (pipes, &&, $(...), redirection) instead of a program with literal arguments
      --timeout duration   max time the remote command may run before it's aborted; 0 disables the cap (Go duration syntax: 30s, 10m, 1h) (default 10m0s)
```

### Options inherited from parent commands

```
      --ci                 If true it indicates that we're used by another tool so don't require any user input!
      --debug              If true it enables DEBUG mode.
      --no-color           Disable ANSI colors (the NO_COLOR env var is also honored).
  -o, --output string      Output format for commands that support it. Accepted: raw (default), json, yml (alias "human").
      --pr                 If true bitrise runs in pull request mode.
  -q, --quiet              Suppress non-error diagnostic messages.
      --theme string       Color theme. Accepted: auto, dark, light, none.
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
```

### SEE ALSO

* [bitrise rde session](bitrise_rde_session.md)	 - Create, list, inspect, and manage RDE sessions

