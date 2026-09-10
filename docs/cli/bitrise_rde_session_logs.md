## bitrise rde session logs

Print a session's warmup or startup logs

### Synopsis

Print the warmup or startup script logs for a session — useful for debugging a
session stuck provisioning or one that came up failed. The log replays from the
start every time you connect.

By default this prints the log captured so far and exits, without waiting for
more output. Pass --follow to keep streaming new output live until you stop it
with Ctrl-C (the backend does not signal end-of-log, so --follow runs until
interrupted), and to wait for the stage to start if it hasn't produced any
logs yet instead of erroring.

warmup logs run once at session creation; startup logs run on every session
start/restart.

--format is rejected — logs stream as raw text, not a single object. Pipe or
redirect as needed; diagnostics go to stderr so a redirect captures only log
text.

```
bitrise rde session logs SESSION_ID --stage warmup|startup [flags]
```

### Examples

```
  bitrise rde session logs SESSION_ID --stage startup
  bitrise rde session logs SESSION_ID --stage warmup
  bitrise rde session logs SESSION_ID --stage startup --follow
  bitrise rde session logs SESSION_ID --stage startup > session.log
```

### Options

```
  -f, --follow          keep streaming until Ctrl-C, waiting for the stage to start if needed
      --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for logs
      --stage string    which logs to show: warmup or startup (required)
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

