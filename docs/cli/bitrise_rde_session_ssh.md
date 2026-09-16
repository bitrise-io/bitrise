## bitrise rde session ssh

Print SSH connection details (command, host, port, user, password) for a session

### Synopsis

Print the SSH connection details for a running session: a ready-to-run ssh
command line, the host, port and user it decomposes into, and the session's
SSH password.

The password is ephemeral and tied to this session. `rde session view` and
`--format json` on other commands intentionally hide it; this command is the
opt-in way to get it — for scp, an SSH tunnel to a device session's ports, or
any tool that is not `rde session exec` (which dials for you and needs none
of this).

Human mode prints the command on the first line and the password on the
second. --password-only prints just the password, so it can be fed to an
SSH_ASKPASS helper without parsing:

  RDE_SSH_PASSWORD="$(bitrise rde session ssh SESSION_ID --password-only)"

--format json emits {address, host, port, user, password, command}.

```
bitrise rde session ssh SESSION_ID [flags]
```

### Examples

```
  bitrise rde session ssh SESSION_ID
  bitrise rde session ssh SESSION_ID --format json
  RDE_SSH_PASSWORD="$(bitrise rde session ssh SESSION_ID --password-only)"
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for ssh
      --password-only   print only the SSH password (for SSH_ASKPASS helpers and scripts)
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

