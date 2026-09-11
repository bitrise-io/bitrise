## bitrise build list

List builds for an app

### Synopsis

List builds for an app.

BUILD_ID's app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

In JSON mode (--format json), next_cursor holds the cursor value for scripting:
  bitrise build list --app my-app-id --format json | jq -r '.next_cursor'

```
bitrise build list [flags]
```

### Examples

```
  bitrise build list --app my-app-id
  bitrise build list --app my-app-id --branch main --status failed
  bitrise build list --app my-app-id --all
```

### Options

```
      --after string                only builds triggered after this time (RFC3339)
      --all                         fetch all pages automatically
      --app string                  app ID (or set BITRISE_APP_ID)
      --before string               only builds triggered before this time (RFC3339)
      --branch string               filter by branch
      --build-number int            filter by build number
      --commit-message string       filter by commit message
      --cursor string               pagination cursor from a previous response
  -f, --format string               Output format. Accepted: raw (default), json, yml
  -h, --help                        help for list
      --limit int                   max items per page (server default if 0)
      --pipeline-build              filter by whether the build is part of a pipeline
      --pull-request-id int         filter by pull request ID
      --sort-by string              ordering: created_at (default) or running_first
      --status string               filter by status: in-progress, success, failed, aborted, aborted-with-success
      --trigger-event-type string   filter by trigger event type: push, pull-request, tag
      --workflow string             filter by workflow
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

