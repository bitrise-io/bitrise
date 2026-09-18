## bitrise rde session restore

Restore a terminated session (re-provisions its VM from the persistent disk)

### Synopsis

Restore a terminated session (re-provisions its VM from the persistent disk).

Restore is asynchronous: by default the command returns while the session is
still "starting". Pass --wait to block until the session finishes provisioning
(mirrors 'session create --wait'); the command exits non-zero if the session
ends in any state other than "running". This lets an unattended caller restore
and then immediately use the session without hand-rolling a poll loop.

```
bitrise rde session restore SESSION_ID [flags]
```

### Options

```
  -f, --format string           Output format. Accepted: raw (default), json, yml
  -h, --help                    help for restore
      --wait                    wait until the session leaves provisioning (running, failed, …) before returning; exits 1 if the final status isn't running
      --wait-timeout duration   max time to wait when --wait is set (Go duration syntax: 30s, 5m, 1h) (default 10m0s)
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

