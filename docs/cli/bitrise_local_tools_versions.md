## bitrise local tools versions

List available versions for a supported tool

### Synopsis

List available versions for a supported tool

TOOL: tool name (e.g. nodejs, golang, ruby)
VERSION_PREFIX: optional version prefix to filter by (e.g. 22, 3.12)

EXAMPLES:
   bitrise tools versions nodejs
   bitrise tools versions nodejs 22
   bitrise tools versions nodejs --format json

```
bitrise local tools versions TOOL [VERSION_PREFIX] [--format FORMAT] [flags]
```

### Options

```
  -f, --format string     Output format. Options: plaintext, json (default "plaintext")
  -h, --help              help for versions
  -p, --provider string   Tool provider to use (asdf/mise) (default: "mise")
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

