## bitrise build

Trigger, list, and inspect builds.

```
bitrise build [flags]
```

### Options

```
  -h, --help   help for build
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
* [bitrise build abort](bitrise_build_abort.md)	 - Abort a running or queued build
* [bitrise build list](bitrise_build_list.md)	 - List builds for an app
* [bitrise build log](bitrise_build_log.md)	 - Print the build log
* [bitrise build trigger](bitrise_build_trigger.md)	 - Start a new build
* [bitrise build view](bitrise_build_view.md)	 - Show details of a single build
* [bitrise build watch](bitrise_build_watch.md)	 - Stream logs for a running build
* [bitrise build yml](bitrise_build_yml.md)	 - Print the bitrise.yml a specific build ran with

