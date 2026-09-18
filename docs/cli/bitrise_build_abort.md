## bitrise build abort

Abort a running or queued build

### Synopsis

Abort a running or queued build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID.

```
bitrise build abort BUILD_ID [flags]
```

### Examples

```
  bitrise build abort abc123 --app my-app-id
  bitrise build abort abc123 --app my-app-id --reason "no longer needed"
  bitrise build abort abc123 --app my-app-id --abort-with-success
```

### Options

```
      --abort-with-success       mark the aborted build as successful
      --app string               app ID (or set BITRISE_APP_ID)
  -f, --format string            Output format. Accepted: raw (default), json, yml
  -h, --help                     help for abort
      --reason string            reason for aborting, recorded on the build and available via the UI/API
      --skip-git-status-report   don't report the abort to the git provider's status API
      --skip-notifications       don't send abort notifications
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

