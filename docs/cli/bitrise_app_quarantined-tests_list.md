## bitrise app quarantined-tests list

List the tests quarantined on an app

```
bitrise app quarantined-tests list [flags]
```

### Examples

```
  bitrise app quarantined-tests list --app my-app-id
  bitrise app quarantined-tests list --app my-app-id --format json
```

### Options

```
      --app string      app ID (or set BITRISE_APP_ID)
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

* [bitrise app quarantined-tests](bitrise_app_quarantined-tests.md)	 - List and manage the tests quarantined on an app

