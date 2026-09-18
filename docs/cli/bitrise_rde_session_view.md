## bitrise rde session view

Show details of a single session

### Synopsis

Show details of a single session.

Pass --watch to poll the session and re-render on every change until you
hit Ctrl-C — useful while waiting for a session to come up. --watch is
incompatible with a machine-readable --format (the contract is a single
object, not a stream); use 'session create --wait' or a polling jq loop
instead.

```
bitrise rde session view SESSION_ID [flags]
```

### Options

```
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for view
      --interval duration   polling interval when --watch is set (Go duration syntax: 1s, 500ms, …) (default 3s)
      --watch               poll the session and re-render on every change until Ctrl-C
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

