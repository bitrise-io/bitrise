## bitrise app view

Show details of a single app

### Synopsis

Show details for a single app identified by its ID.

APP_ID falls back to --app, then $BITRISE_APP_ID, then $BITRISE_APP_SLUG
(injected inside a Bitrise build), then the app_id saved by 'bitrise app
create' or 'bitrise config set app_id', when omitted.

```
bitrise app view [APP_ID] [flags]
```

### Examples

```
  bitrise app view stub-app-aaa
  bitrise app view stub-app-aaa --format json
  bitrise app view stub-app-aaa --web
  BITRISE_APP_ID=stub-app-aaa bitrise app view
```

### Options

```
      --app string      app ID to view (or set BITRISE_APP_ID); overridden by the positional argument
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for view
      --web             open the app page in the browser instead of printing
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

* [bitrise app](bitrise_app.md)	 - List, inspect, and manage apps.

