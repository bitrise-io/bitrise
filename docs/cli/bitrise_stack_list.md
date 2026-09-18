## bitrise stack list

List available stacks and their machine configurations

### Synopsis

List all available stacks with their OS, status, and version information.

When a workspace resolves (via --workspace, its env var, or a configured
default), returns stacks available for that workspace, including any custom
stacks configured for it. Otherwise returns globally available stacks.

```
bitrise stack list [flags]
```

### Examples

```
  bitrise stack list
  bitrise stack list --workspace my-workspace-id
  bitrise stack list --format json
```

### Options

```
  -f, --format string      Output format. Accepted: raw (default), json, yml
  -h, --help               help for list
      --workspace string   workspace ID for workspace-specific stacks (including custom stacks), or set BITRISE_WORKSPACE_ID / default_workspace_id
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

* [bitrise stack](bitrise_stack.md)	 - Manage build stacks.

