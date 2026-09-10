## bitrise local tools latest

Query the latest version of a tool

### Synopsis

Query the latest version of a tool, optionally matching a version prefix.

By default, queries latest available release. Use :installed suffix for latest installed version.

EXAMPLES:
   bitrise tools latest nodejs
   bitrise tools latest nodejs 20
   bitrise tools latest python 3.12:installed
   bitrise tools latest ruby installed

```
bitrise local tools latest [--provider PROVIDER] [--format FORMAT] <TOOL> [VERSION[:SUFFIX]] [flags]
```

### Options

```
  -f, --format string     Output format of the env vars that activate installed tools. Options: plaintext, json, bash (default "plaintext")
  -h, --help              help for latest
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

