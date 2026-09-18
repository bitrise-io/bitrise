## bitrise local tools catalog

List officially supported tools

### Synopsis

List officially supported tools

EXAMPLES:
   bitrise tools catalog
   bitrise tools catalog --format json

```
bitrise local tools catalog [--format FORMAT] [flags]
```

### Options

```
  -f, --format string   Output format. Options: plaintext, json (default "plaintext")
  -h, --help            help for catalog
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

