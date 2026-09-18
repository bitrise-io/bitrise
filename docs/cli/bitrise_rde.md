## bitrise rde

Manage Bitrise Remote Dev Environments (sessions, templates, …)

### Synopsis

Manage Bitrise Remote Dev Environments — sessions, templates, saved inputs,
and the machine catalog (stacks, machine types).

If --workspace isn't resolved from a flag, env var, or configured default,
and you belong to more than one workspace, you're prompted to pick one on a
terminal, or shown a sorted list of workspaces to choose from via --workspace
otherwise.

Saved inputs are user-scoped, though — they do not require --workspace, and
the 'saved-input' subcommand does not accept it.

```
bitrise rde [flags]
```

### Examples

```
  bitrise rde session list --workspace WORKSPACE_ID
  bitrise rde session list --format json
  bitrise rde machine-type list --stack osx-xcode-16.0.x-edge
```

### Options

```
  -h, --help   help for rde
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

* [bitrise](bitrise.md)	 - Bitrise Automations Workflow Runner
* [bitrise rde claude](bitrise_rde_claude.md)	 - Create an RDE session and attach to Claude Code
* [bitrise rde machine-type](bitrise_rde_machine-type.md)	 - List machine types compatible with a given stack
* [bitrise rde saved-input](bitrise_rde_saved-input.md)	 - Manage saved inputs (reusable credentials/values)
* [bitrise rde session](bitrise_rde_session.md)	 - Create, list, inspect, and manage RDE sessions
* [bitrise rde stack](bitrise_rde_stack.md)	 - List machine stacks available to the workspace
* [bitrise rde template](bitrise_rde_template.md)	 - List and inspect RDE templates
* [bitrise rde usage](bitrise_rde_usage.md)	 - Show the workspace's active session and resource usage

