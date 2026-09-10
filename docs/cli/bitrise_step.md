## bitrise step

Manage steps.

```
bitrise step [flags]
```

### Options

```
  -h, --help   help for step
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
* [bitrise step inputs](bitrise_step_inputs.md)	 - List inputs of a step version
* [bitrise step list-cached](bitrise_step_list-cached.md)	 - List all the cached steps
* [bitrise step preload](bitrise_step_preload.md)	 - Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps.
* [bitrise step search](bitrise_step_search.md)	 - Find steps by name, description, or tags
* [bitrise step share](bitrise_step_share.md)	 - Publish your step.

