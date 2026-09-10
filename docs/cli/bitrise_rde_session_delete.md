## bitrise rde session delete

Permanently delete a session

### Synopsis

Permanently delete a session.

The session must already be terminated or failed — delete rejects a session
that's still running or terminating. Use 'terminate --wait' first, or
'delete-terminated' to sweep every already-terminated session at once.

```
bitrise rde session delete SESSION_ID [flags]
```

### Options

```
  -h, --help   help for delete
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

