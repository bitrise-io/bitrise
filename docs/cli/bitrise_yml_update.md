## bitrise yml update

Upload a new bitrise.yml to Bitrise

### Synopsis

Upload a new bitrise.yml configuration to Bitrise for an app.

Reads from --file if provided, otherwise reads from stdin. Pass --file - to
read from stdin explicitly.

Note: if the app is configured to read its bitrise.yml from the repository,
this command succeeds but the change will not affect builds.

Inside a Bitrise build, omitting --app targets the app the build runs for,
overwriting its own stored configuration. Always pass --app when updating a
different app from a build.

Bitrise stores the configuration as structured data rather than as the file
you upload, so comments, key order and YAML anchors are not preserved: a
later 'bitrise yml get' returns an equivalent, reformatted document.

```
bitrise yml update [flags]
```

### Examples

```
  bitrise yml update --app my-app-id --file bitrise.yml
  cat bitrise.yml | bitrise yml update --app my-app-id
  bitrise yml update --app my-app-id < bitrise.yml
```

### Options

```
      --app string    app ID to update the bitrise.yml for (or set BITRISE_APP_ID; inside a build, defaults to the app the build runs for)
  -f, --file string   path to the bitrise.yml file, or - for stdin (reads from stdin if omitted)
  -h, --help          help for update
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

