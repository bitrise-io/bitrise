## bitrise rde saved-input create

Create a new saved input

### Synopsis

Create a new saved input.

The value can be supplied three ways:
  --value VALUE   use VALUE literally (pass --value - to store a literal dash)
  --value-stdin   read the value from stdin without prompting; keeps secrets
                  out of shell history
  neither         prompt for the value interactively; input is masked when
                  stdin is a terminal

```
bitrise rde saved-input create [flags]
```

### Examples

```
  bitrise rde saved-input create --key repo-name --value my-app
  echo -n "ghp_xxx" | bitrise rde saved-input create --key gh-token --value-stdin --secret
  bitrise rde saved-input create --key gh-token --secret   # prompts for the value
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for create
      --key string      saved-input key (required)
      --secret          encrypt value at rest; the value will be masked in subsequent reads
      --value string    value to store (literal)
      --value-stdin     read the value from stdin without prompting
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

