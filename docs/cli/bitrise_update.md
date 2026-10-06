## bitrise update

Updates the Bitrise CLI.

### Synopsis

Updates the Bitrise CLI to the newest release inside the current major version.

A newer major version is never installed on its own, because it can contain breaking
changes. It is reported instead, together with the command that installs it.

```
bitrise update [flags]
```

### Options

```
  -h, --help             help for update
      --version string   exact version to install, instead of the newest one in the current major version - only for GitHub release page installations.
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

* [bitrise](bitrise.md)	 - Bitrise Automations Workflow Runner

