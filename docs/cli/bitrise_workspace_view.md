## bitrise workspace view

Show details of a single workspace

### Synopsis

Show details for a single workspace identified by its ID or name.

WORKSPACE_ID falls back to --workspace, then $BITRISE_WORKSPACE_ID, then the
default_workspace_id config key. With none of them set, your only workspace is
used, or you pick one interactively when you have several.

```
bitrise workspace view [WORKSPACE_ID] [flags]
```

### Examples

```
  bitrise workspace view my-workspace-id
  bitrise workspace view "My Workspace"
  bitrise workspace view my-workspace-id --format json
  bitrise workspace view my-workspace-id --web
```

### Options

```
  -f, --format string      Output format. Accepted: raw (default), json, yml
  -h, --help               help for view
      --web                open the workspace page in the browser instead of printing
      --workspace string   workspace ID or name to view (or set BITRISE_WORKSPACE_ID / default_workspace_id); overridden by the positional argument
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

* [bitrise workspace](bitrise_workspace.md)	 - List and inspect workspaces.

