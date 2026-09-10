## bitrise local tools install

Install a specific tool version

### Synopsis

Install a specific version of a tool using the configured tool provider.

TOOL: tool name (e.g., nodejs, ruby, python, go, etc.)
VERSION: specific version (20.10.0), prefix (22), latest, or installed.

EXAMPLES:
   bitrise tools install nodejs 20.10.0
   bitrise tools install nodejs 22:latest
   eval "$(bitrise tools install ruby 3.2.0 --format bash)"  # activate in shell

```
bitrise local tools install [--provider PROVIDER] [--format FORMAT] <TOOL> <VERSION>[:SUFFIX] [flags]
```

### Options

```
  -f, --format string     Output format of the env vars that activate installed tools. Options: plaintext, json, bash (default "plaintext")
  -h, --help              help for install
  -p, --provider string   Tool provider to use (asdf/mise). If not specified, uses the default.
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

