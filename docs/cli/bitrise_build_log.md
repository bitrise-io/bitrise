## bitrise build log

Print the build log

### Synopsis

Print the log output for a single build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

--wait waits for the build to finish before printing the log — useful when
the build is still in-progress. Ctrl-C detaches without affecting the
running build.

Output is always raw text — this command has no --format flag.

```
bitrise build log BUILD_ID [flags]
```

### Examples

```
  bitrise build log abc123 --app my-app-id
  bitrise build log abc123 --app my-app-id --wait
  bitrise build log abc123 --app my-app-id --wait --interval 10s
  bitrise build log abc123 --app my-app-id > build.log
```

### Options

```
      --app string          app ID (or set BITRISE_APP_ID)
  -h, --help                help for log
      --interval duration   polling interval when --wait is active (default 3s)
      --wait                wait for the build to finish before printing the log
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

