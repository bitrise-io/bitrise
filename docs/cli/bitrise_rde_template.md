## bitrise rde template

List and inspect RDE templates

### Synopsis

List and inspect RDE templates.

Commands that take a TEMPLATE_ID also accept a template name — it's resolved
to an ID for you. Names aren't unique, so if more than one template shares the
name the command errors and lists the candidate IDs to pick from.

```
bitrise rde template [flags]
```

### Options

```
  -h, --help               help for template
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
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

* [bitrise rde](bitrise_rde.md)	 - Manage Bitrise Remote Dev Environments (sessions, templates, …)
* [bitrise rde template create](bitrise_rde_template_create.md)	 - Create a new RDE template from a JSON spec file
* [bitrise rde template delete](bitrise_rde_template_delete.md)	 - Delete an RDE template
* [bitrise rde template list](bitrise_rde_template_list.md)	 - List RDE templates in the workspace
* [bitrise rde template update](bitrise_rde_template_update.md)	 - Update an existing RDE template from a JSON spec file
* [bitrise rde template view](bitrise_rde_template_view.md)	 - Show details of a single template

