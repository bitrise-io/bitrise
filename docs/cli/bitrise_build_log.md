## bitrise build log

Print the build log

### Synopsis

Print the log output for a single build.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

--step-status and --step print the logs of the matching steps only, each
under a header with its workflow, status and duration. --step takes the step
IDs "bitrise build steps" lists. Step logs are available once the build has
finished and its log is archived, a little after; until then only the steps
are listed, without logs.

--tail N keeps the last N lines of the log, or of each step's log.

--wait waits for the build to finish before printing the log — useful when
the build is still in-progress. With --step-status or --step it also waits
for the log to be archived. Ctrl-C detaches without affecting the running build, and
exits non-zero because no log was printed.

```
bitrise build log BUILD_ID [flags]
```

### Examples

```
  bitrise build log abc123 --app my-app-id
  bitrise build log abc123 --app my-app-id --tail 100
  bitrise build log abc123 --app my-app-id --step-status failed
  bitrise build log abc123 --app my-app-id --step 3f2a1c9e-... --format json
  bitrise build log abc123 --app my-app-id --wait --step-status failed --tail 50
  bitrise build log abc123 --app my-app-id > build.log
```

### Options

```
      --app string           app ID (or set BITRISE_APP_ID)
  -f, --format string        Output format. Accepted: raw (default), json, yml
  -h, --help                 help for log
      --interval duration    polling interval when --wait is active (default 3s)
      --step strings         only the log of the step with this ID, as listed by "bitrise build steps" (repeatable)
      --step-status string   only the logs of steps with this status: not_started, started, success, failed, failed_skippable, skipped, skipped_with_run_if, preparation_failed, aborted, aborted_with_custom_timeout, aborted_with_no_output. failed matches every status that fails the build: failed, preparation_failed, aborted_with_custom_timeout, aborted_with_no_output
      --tail int             only the last N lines of the log, or of each step's log
      --wait                 wait for the build to finish before printing the log
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

