## bitrise rde saved-input update

Update a saved input's value and/or secret flag

### Synopsis

Update a saved input's value and/or secret flag.

Pass --value VALUE to set a new value, or --value-stdin to read it from stdin
without prompting (keeps secrets out of shell history). Omit both to change only
the --secret flag.

```
bitrise rde saved-input update SAVED_INPUT_ID [flags]
```

### Examples

```
  bitrise rde saved-input update ID --value new-value
  echo -n "ghp_xxx" | bitrise rde saved-input update ID --value-stdin --secret
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for update
      --secret          set/unset the secret flag
      --value string    new value (literal)
      --value-stdin     read the new value from stdin without prompting
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

* [bitrise rde saved-input](bitrise_rde_saved-input.md)	 - Manage saved inputs (reusable credentials/values)

