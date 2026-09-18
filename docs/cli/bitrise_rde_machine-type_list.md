## bitrise rde machine-type list

List machine types compatible with a given stack

### Synopsis

List machine types compatible with the stack given by --stack.

Each machine type is offered by one or more clusters. A CLUSTER column is
shown when any machine type is offered by more than one cluster for the
selected stack — pass that name as --cluster to 'rde session create' to
pin a target.

```
bitrise rde machine-type list --stack STACK_ID [flags]
```

### Examples

```
  bitrise rde machine-type list --stack osx-xcode-16.0.x-edge
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for list
      --stack string    stack ID to list compatible machine types for (required)
```

### Options inherited from parent commands

```
      --ci                 If true it indicates that we're used by another tool so don't require any user input!
      --debug              If true it enables DEBUG mode.
      --no-color           Disable ANSI colors (the NO_COLOR env var is also honored).
  -o, --output string      Output format for commands that support it. Accepted: raw (default), json, yml (alias "human").
      --pr                 If true bitrise runs in pull request mode.
  -q, --quiet              Suppress non-error diagnostic messages.
      --theme string       Color theme. Accepted: auto, dark, light, none.
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
```

### SEE ALSO

* [bitrise rde machine-type](bitrise_rde_machine-type.md)	 - List machine types compatible with a given stack

