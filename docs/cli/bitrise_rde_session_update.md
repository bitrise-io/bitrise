## bitrise rde session update

Update a session's name, description, auto-terminate duration, or labels

### Synopsis

Update a session's name, description, auto-terminate duration, or labels.

Labels change incrementally: --label key=value upserts one label (an existing
key is overwritten, other keys are left untouched) and --unset-label key
removes one; both are repeatable. Removing a key the session doesn't have is
a no-op.

```
bitrise rde session update SESSION_ID [flags]
```

### Examples

```
  bitrise rde session update SESSION_ID --name new-name
  bitrise rde session update SESSION_ID --auto-terminate-minutes 0
  bitrise rde session update SESSION_ID --label branch=main --unset-label wip
```

### Options

```
      --auto-terminate-minutes int   auto-terminate duration in minutes; 0 disables. Resets the deadline to now + minutes.
      --description string           new session description
  -f, --format string                Output format. Accepted: raw (default), json, yml
  -h, --help                         help for update
  -l, --label stringArray            label to set on the session as key=value (repeatable; merged into the existing labels)
      --name string                  new session name
      --unset-label stringArray      label key to remove from the session (repeatable; unknown keys are ignored)
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

* [bitrise rde session](bitrise_rde_session.md)	 - Create, list, inspect, and manage RDE sessions

