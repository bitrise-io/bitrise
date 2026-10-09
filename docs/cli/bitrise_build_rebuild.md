## bitrise build rebuild

Start a new build with the parameters of a finished build

### Synopsis

Start a new build of a finished build with the same parameters and workflow.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID.

--wait and --watch behave as in "bitrise build trigger".

```
bitrise build rebuild BUILD_ID [flags]
```

### Examples

```
  bitrise build rebuild abc123 --app my-app-id
  bitrise build rebuild abc123 --app my-app-id --watch
  bitrise build rebuild abc123 --app my-app-id --remote-access
  bitrise build rebuild abc123 --app my-app-id --wait --format json
```

### Options

```
      --app string          app ID (or set BITRISE_APP_ID)
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for rebuild
      --interval duration   polling interval when --wait or --watch is active (default 3s)
      --remote-access       start the new build with remote access enabled (needs remote access permission on the app)
      --wait                block until the build finishes without streaming logs (exit code reflects build outcome)
      --watch               wait for the build to finish, showing progress (exit code reflects build outcome)
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

