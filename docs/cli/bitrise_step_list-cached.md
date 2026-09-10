## bitrise step list-cached

List all the cached steps

```
bitrise step list-cached [flags]
```

### Options

```
  -h, --help                 help for list-cached
      --maintainer string    Maintainer of the steps to list or preload (default "bitrise")
      --steplib-url string   URL of the steplib to list or preload steps from (default "https://github.com/bitrise-io/bitrise-steplib.git")
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

