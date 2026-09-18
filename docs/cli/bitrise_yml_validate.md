## bitrise yml validate

Validates a specified bitrise config.

### Synopsis

Validate a bitrise.yml (and, if present, an inventory/secrets file).

By default, bitrise.yml is read from the current directory; use --config to
point elsewhere, --config-base64 to pass it inline, or --config - to read it
from stdin.

Validation runs online (against the Bitrise API, which also checks
app-specific things like stacks, machine types, and license pools when --app
is given) whenever you're authenticated; pass --offline to force the local
schema-only check instead, or --app to enable the app-specific checks
explicitly. Falls back to the local check automatically if the online
attempt can't complete.

```
bitrise yml validate [flags]
```

### Examples

```
  bitrise yml validate
  bitrise yml validate --config ./ci/bitrise.yml
  bitrise yml validate --config - < bitrise.yml
  bitrise yml validate --offline
  bitrise yml validate --app my-app-id --format json
```

### Options

```
      --app string                app ID to validate against (enables app-specific checks: stacks, machine types, license pools; inside a build, defaults to the app the build runs for)
  -c, --config string             Path where the workflow config file is located.
      --config-base64 string      base64 encoded config data.
      --format string             Output format. Accepted: raw (default), json, yml.
  -h, --help                      help for validate
  -i, --inventory string          Path of the inventory file.
      --inventory-base64 string   base64 encoded inventory data.
      --offline                   Skip online validation even if authenticated; use only the local schema check.
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

