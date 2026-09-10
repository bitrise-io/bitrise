## bitrise step search

Find steps by name, description, or tags

### Synopsis

Find steps for use in workflows or step bundles.

Returns only the latest, non-deprecated version of each matching step.

Valid categories:
  build, code-sign, test, deploy, notification, access-control,
  artifact-info, installer, dependency, utility, security

Valid maintainers:
  bitrise   official Bitrise steps
  verified  verified community steps
  community all community steps

```
bitrise step search QUERY [flags]
```

### Examples

```
  bitrise step search clone
  bitrise step search deploy --category deploy --maintainer bitrise
  bitrise step search npm --format json
```

### Options

```
      --category stringArray     filter by category (may be repeated)
  -f, --format string            Output format. Accepted: raw (default), json, yml
  -h, --help                     help for search
      --maintainer stringArray   filter by maintainer: bitrise, verified, community (may be repeated)
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

