## bitrise config get

Print the value of a single config key

### Synopsis

Print the value of one config key from the global config file.

Valid keys: api_base_url, web_base_url, rde_api_base_url, app_id, default_workspace_id, output, theme

```
bitrise config get KEY [flags]
```

### Options

```
  -h, --help   help for get
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

