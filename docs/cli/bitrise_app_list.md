## bitrise app list

List apps the authenticated user can access

### Synopsis

List all apps the authenticated user can access.

When a workspace resolves (via --workspace, its env var, or a configured
default), only apps owned by that workspace are returned. Otherwise every
app the authenticated user can access, across all workspaces, is returned.

In JSON mode (--format json), next_cursor holds the cursor value for scripting:
  bitrise app list --format json | jq -r '.next_cursor'

```
bitrise app list [flags]
```

### Examples

```
  bitrise app list
  bitrise app list --workspace acme
  bitrise app list --all
  bitrise app list --format json | jq -r '.next_cursor'
  bitrise app list --project-type ios --limit 100
```

### Options

```
      --all                   fetch all pages automatically
      --cursor string         pagination cursor from a previous response
  -f, --format string         Output format. Accepted: raw (default), json, yml
  -h, --help                  help for list
      --limit int             max items per page (server default if 0)
      --project-type string   filter by project type (ios, android, ...)
      --sort-by string        ordering accepted by the API (created_at, last_build_at)
      --title string          filter apps by title
      --workspace string      only list apps owned by this workspace, or set BITRISE_WORKSPACE_ID / default_workspace_id
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

* [bitrise app](bitrise_app.md)	 - List, inspect, and manage apps.

