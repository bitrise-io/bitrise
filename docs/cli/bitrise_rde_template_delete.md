## bitrise rde template delete

Delete an RDE template

### Synopsis

Delete an RDE template (soft-delete server-side). Existing sessions
created from this template keep working — they reference a snapshot — but
the template can no longer be selected for new sessions.

```
bitrise rde template delete TEMPLATE_ID [flags]
```

### Options

```
  -h, --help   help for delete
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

* [bitrise rde template](bitrise_rde_template.md)	 - List and inspect RDE templates

