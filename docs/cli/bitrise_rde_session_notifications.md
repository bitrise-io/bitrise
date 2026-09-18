## bitrise rde session notifications

List notifications emitted by a session

### Synopsis

List notifications a session has emitted (agent stop, permission prompt,
idle, …). Use --since with the timestamp of the newest notification you've
seen to poll for new events incrementally.

```
bitrise rde session notifications SESSION_ID [flags]
```

### Examples

```
  bitrise rde session notifications SESSION_ID
  bitrise rde session notifications SESSION_ID --since 2026-05-27T10:00:00Z --limit 100
  bitrise rde session notifications SESSION_ID --order asc
```

### Options

```
      --before string   only notifications created before this RFC3339 timestamp (exclusive)
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for notifications
      --limit int       max notifications to return (server default 50, max 100)
      --order string    sort order: asc (oldest first) or desc (newest first); server default is desc
      --since string    only notifications created after this RFC3339 timestamp (exclusive)
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

