## bitrise rde stack

List machine stacks available to the workspace

```
bitrise rde stack [flags]
```

### Options

```
  -h, --help               help for stack
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
* [bitrise rde stack list](bitrise_rde_stack_list.md)	 - List machine stacks

