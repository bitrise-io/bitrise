## bitrise rde usage

Show the workspace's active session and resource usage

### Synopsis

Show a point-in-time snapshot of the workspace's active remote dev sessions:
session counts and vCPU/memory totals split by OS, workspace-wide and per user.

This reports sessions currently consuming resources; it is not a historical or
billing-period report. Requires the workspace's billing-view permission
(workspace owners and billing-managing custom roles).

```
bitrise rde usage [flags]
```

### Examples

```
  bitrise rde usage
  bitrise rde usage --format json | jq '.totals'
```

### Options

```
  -f, --format string      Output format. Accepted: raw (default), json, yml
  -h, --help               help for usage
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
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

