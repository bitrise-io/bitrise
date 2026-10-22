## bitrise build steps

List a build's steps with their status and failure or skip reasons

### Synopsis

List the steps a build ran, across all its workflows, in execution order,
with each step's status and timing. Failed steps include the errors they
reported, and skipped steps the reason they were skipped.

While the build is still running, or its log is still being archived, only
the steps that have started so far are listed.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

```
bitrise build steps BUILD_ID [flags]
```

### Examples

```
  bitrise build steps abc123 --app my-app-id
  bitrise build steps abc123 --app my-app-id --status failed
  bitrise build steps abc123 --app my-app-id --format json
```

### Options

```
      --app string      app ID (or set BITRISE_APP_ID)
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for steps
      --status string   only steps with this status: not_started, started, success, failed, failed_skippable, skipped, skipped_with_run_if, preparation_failed, aborted, aborted_with_custom_timeout, aborted_with_no_output. failed matches every status that fails the build: failed, preparation_failed, aborted_with_custom_timeout, aborted_with_no_output
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

* [bitrise build](bitrise_build.md)	 - Trigger, list, and inspect builds.

