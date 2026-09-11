## bitrise rde session diff

Compare a session's template snapshot with the current template

### Synopsis

Show what changed between the template config snapshotted at the session's
creation time and the template's current config. Most useful when a session
reports template_outdated=true.

Lists which template variable keys changed (values are never exposed) and
the simple per-field differences (stack, machine type, scripts, working
directory). When the template was deleted, only the snapshot is shown.

```
bitrise rde session diff SESSION_ID [flags]
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for diff
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

