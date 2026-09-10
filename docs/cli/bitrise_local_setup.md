## bitrise local setup

Setup the current host. Install every required tool to run Workflows.

```
bitrise local setup [flags]
```

### Options

```
      --clean       Removes bitrise's workdir before setup.
  -h, --help        help for setup
      --minimal     Only installs the required tools for running in CI mode.
      --no-update   Skip updating core tools (stepman/envman) and plugins if they are already installed, even if outdated (or set BITRISE_SETUP_NO_UPDATE=true).
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

