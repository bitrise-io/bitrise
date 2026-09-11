## bitrise build view

Show details of a single build

### Synopsis

Show details for a single build identified by its ID.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

```
bitrise build view BUILD_ID [flags]
```

### Examples

```
  bitrise build view abc123 --app my-app-id
  bitrise build view abc123 --app my-app-id --format json
  bitrise build view abc123 --app my-app-id --web
```

### Options

```
      --app string      app ID (or set BITRISE_APP_ID)
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for view
      --web             open the build page in the browser instead of printing
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

* [bitrise build](bitrise_build.md)	 - Trigger, list, and inspect builds.

