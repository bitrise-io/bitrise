## bitrise local

Run and manage Bitrise workflows on the local host.

```
bitrise local [flags]
```

### Options

```
  -h, --help   help for local
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
* [bitrise local init](bitrise_local_init.md)	 - Init bitrise config.
* [bitrise local run](bitrise_local_run.md)	 - Runs a specified Workflow.
* [bitrise local setup](bitrise_local_setup.md)	 - Setup the current host. Install every required tool to run Workflows.
* [bitrise local tools](bitrise_local_tools.md)	 - Manage available tools from inside the workflow.
* [bitrise local workflows](bitrise_local_workflows.md)	 - List of available workflows in config.

