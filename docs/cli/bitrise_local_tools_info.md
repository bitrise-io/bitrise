## bitrise local tools info

Show information about installed or active tools.

### Synopsis

Show information about installed or active tools.

Show installed tool versions. Use --active to show only tools currently active in the shell context.

EXAMPLES:
   bitrise tools info
   bitrise tools info --active
   bitrise tools info --active --format json

```
bitrise local tools info [--active] [--format FORMAT] [flags]
```

### Options

```
  -a, --active          Show only currently active tools in the shell context (based on config files in current directory)
  -f, --format string   Output format of the env vars that activate installed tools. Options: plaintext, json, bash (default "plaintext")
  -h, --help            help for info
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

* [bitrise local tools](bitrise_local_tools.md)	 - Manage available tools from inside the workflow.

