## bitrise rde session list

List RDE sessions in the workspace

### Synopsis

List RDE sessions in the workspace.

Filter by labels with --label-selector key=value (repeatable; selectors are
exact matches and are ANDed, at most 8 per request).

By default every session the authenticated user has in the workspace is
listed. Pass --scope workspace for sessions owned by the workspace itself
rather than by a user (for example sessions spawned by workspace device
preview links) — every workspace member sees the same list. The owning user
or workspace is reported via the owner_type and owner_id fields in
--format json/yml.

The session list comes from the backend in arbitrary order; the CLI does
not paginate (the API doesn't paginate this endpoint either).

```
bitrise rde session list [flags]
```

### Examples

```
  bitrise rde session list
  bitrise rde session list --workspace my-workspace
  bitrise rde session list --scope workspace
  bitrise rde session list -l team=mobile -l branch=main
  bitrise rde session list --format json | jq '.items[].id'
```

### Options

```
  -f, --format string                Output format. Accepted: raw (default), json, yml
  -h, --help                         help for list
  -l, --label-selector stringArray   only sessions whose labels match key=value exactly (repeatable; multiple selectors must all match)
      --scope string                 which sessions to list: mine (sessions you created) or workspace (sessions owned by the workspace itself, visible to every member) (default "mine")
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

