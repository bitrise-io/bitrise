## bitrise local workflows

List of available workflows in config.

```
bitrise local workflows [flags]
```

### Examples

```
  bitrise workflows
  bitrise workflows --minimal
  bitrise workflows --id-only
  bitrise workflows --format json
```

### Options

```
  -c, --config string          Path where the workflow config file is located.
      --config-base64 string   base64 encoded config data.
      --format string          Output format. Accepted: raw, json.
  -h, --help                   help for workflows
      --id-only                Print workflow ids only (mutually exclusive with --minimal).
      --minimal                Print summary of workflows only (mutually exclusive with --id-only).
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

