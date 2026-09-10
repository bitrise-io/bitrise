## bitrise build watch

Stream logs for a running build

### Synopsis

Wait for the build to finish, then exit with a status reflecting the
build outcome (0 = success, non-zero = failed or aborted).

With --format json/yml, or when stdout isn't a terminal, build logs stream to
it as plain text (to stderr instead, in --format json/yml, so stdout stays
pipeable and carries only the final build record). On a terminal with the
default raw format, an interactive status display is shown instead of raw
log lines.

Ctrl-C detaches the CLI without affecting the running build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

```
bitrise build watch BUILD_ID [flags]
```

### Examples

```
  bitrise build watch abc123 --app my-app-id
  bitrise build watch abc123 --app my-app-id --interval 5s
  bitrise build watch abc123 --app my-app-id --format json
```

### Options

```
      --app string          app ID (or set BITRISE_APP_ID)
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for watch
      --interval duration   log polling interval (default 3s)
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

