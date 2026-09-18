## bitrise config list

List the values currently saved in the global config file

### Synopsis

List the values currently saved in the global config file.

This shows what is stored, not what every command will resolve: the
BITRISE_WEB_BASE_URL, BITRISE_RDE_API_BASE_URL, BITRISE_APP_ID, BITRISE_APP_SLUG
and BITRISE_WORKSPACE_ID environment variables, and an app_id or
default_workspace_id pinned by a per-directory .bitrise-cli.yml, all take
precedence at runtime. Inside a Bitrise build, BITRISE_APP_SLUG and
BITRISE_WORKSPACE_ID are always set.

```
bitrise config list [flags]
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for list
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

* [bitrise config](bitrise_config.md)	 - Manage CLI configuration (defaults persisted to a YAML file).

