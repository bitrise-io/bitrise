## bitrise local init

Init bitrise config.

### Synopsis

Interactively create a bitrise.yml (and secrets file) for the current project.

This delegates to the 'init' plugin. If it's not installed yet, 'bitrise
setup' is run automatically first to install it, then init is retried.

```
bitrise local init [flags]
```

### Options

```
  -h, --help      help for init
      --minimal   creates empty bitrise config and secrets
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

* [bitrise local](bitrise_local.md)	 - Run and manage Bitrise workflows on the local host.

