## bitrise yml get

Print the bitrise.yml stored on Bitrise

### Synopsis

Print the bitrise.yml configuration stored on Bitrise for an app.

When --build is provided, prints the bitrise.yml that a specific build ran with
instead of the app's current stored configuration.

```
bitrise yml get [flags]
```

### Examples

```
  bitrise yml get --app my-app-id
  bitrise yml get --app my-app-id --build abc123
  bitrise yml get --app my-app-id --format json
```

### Options

```
      --app string      app ID to retrieve the bitrise.yml for (or set BITRISE_APP_ID; inside a build, defaults to the app the build runs for)
      --build string    build ID to retrieve the yml for
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for get
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

* [bitrise yml](bitrise_yml.md)	 - Work with bitrise.yml files.

