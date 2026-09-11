## bitrise rde session terminate

Terminate a running session (preserves it for later restart)

### Synopsis

Terminate a running session (preserves it for later restart).

Terminate is asynchronous: by default the command returns while the session
is still "terminating". Pass --wait to block until the session settles into a
terminal state ("terminated" or "failed"). This is what makes a
'terminate --wait && delete' pipeline reliable — delete rejects any session
that isn't yet terminated or failed.

```
bitrise rde session terminate SESSION_ID [flags]
```

### Options

```
  -f, --format string           Output format. Accepted: raw (default), json, yml
  -h, --help                    help for terminate
      --wait                    block until the session settles into a terminal state (terminated/failed) before returning; makes 'terminate --wait && delete' reliable
      --wait-timeout duration   max time to wait when --wait is set (Go duration syntax: 30s, 5m, 1h) (default 10m0s)
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

