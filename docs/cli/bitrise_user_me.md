## bitrise user me

Show the currently authenticated user

### Synopsis

Show the profile of the user whose token is in use.

The token comes from the BITRISE_TOKEN environment variable, or from auth.yaml
as written by 'bitrise auth login' — run 'bitrise auth status' to confirm which
source is active.

```
bitrise user me [flags]
```

### Examples

```
  bitrise user me
  bitrise user me --format json
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for me
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

* [bitrise user](bitrise_user.md)	 - Create and manage your Bitrise account.

