## bitrise step share

Publish your step.

### Synopsis

Publish a step to a step library, guiding you through the full flow.

Run bare (no subcommand), this runs the whole interactive publishing wizard.
The subcommands (start, create, audit, finish) let you drive the same steps
individually instead — useful for scripting, or resuming after one step
failed.

```
bitrise step share [flags]
```

### Options

```
  -h, --help   help for share
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

* [bitrise step](bitrise_step.md)	 - Manage steps.
* [bitrise step share audit](bitrise_step_share_audit.md)	 - Validates the step collection.
* [bitrise step share create](bitrise_step_share_create.md)	 - Create your change - add it to your own copy of the collection.
* [bitrise step share finish](bitrise_step_share_finish.md)	 - Finish up.
* [bitrise step share start](bitrise_step_share_start.md)	 - Preparations for publishing.

