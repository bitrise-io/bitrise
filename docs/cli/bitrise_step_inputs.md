## bitrise step inputs

List inputs of a step version

### Synopsis

List the inputs (and their defaults) for a given step version.

STEP_REF must include an exact version: step_id@version

```
bitrise step inputs STEP_REF [flags]
```

### Examples

```
  bitrise step inputs git-clone@8.3.1
  bitrise step inputs git-clone@8.3.1 --format json
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for inputs
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

* [bitrise step](bitrise_step.md)	 - Manage steps.

