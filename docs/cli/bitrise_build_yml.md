## bitrise build yml

Print the bitrise.yml a specific build ran with

### Synopsis

Print the bitrise.yml configuration that a specific build ran with.

This is a shortcut for "bitrise yml get --app ID --build BUILD_ID".

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

```
bitrise build yml BUILD_ID [flags]
```

### Examples

```
  bitrise build yml abc123 --app my-app-id
  bitrise build yml abc123 --app my-app-id --format json
```

### Options

```
      --app string      app ID (or set BITRISE_APP_ID)
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for yml
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

