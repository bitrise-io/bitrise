## bitrise rde session

Create, list, inspect, and manage RDE sessions

### Synopsis

Create, list, inspect, and manage RDE sessions.

Commands that take a SESSION_ID also accept a session name — it's resolved to
an ID for you. Names aren't unique, so if more than one session shares the name
the command errors and lists the candidate IDs to pick from.

```
bitrise rde session [flags]
```

### Options

```
  -h, --help               help for session
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
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

* [bitrise rde](bitrise_rde.md)	 - Manage Bitrise Remote Dev Environments (sessions, templates, …)
* [bitrise rde session create](bitrise_rde_session_create.md)	 - Create a new RDE session
* [bitrise rde session delete](bitrise_rde_session_delete.md)	 - Permanently delete a session
* [bitrise rde session delete-terminated](bitrise_rde_session_delete-terminated.md)	 - Permanently delete every terminated session in the workspace
* [bitrise rde session diff](bitrise_rde_session_diff.md)	 - Compare a session's template snapshot with the current template
* [bitrise rde session download](bitrise_rde_session_download.md)	 - Download a file or directory from a session
* [bitrise rde session exec](bitrise_rde_session_exec.md)	 - Run a command on a session over SSH
* [bitrise rde session list](bitrise_rde_session_list.md)	 - List RDE sessions in the workspace
* [bitrise rde session logs](bitrise_rde_session_logs.md)	 - Print a session's warmup or startup logs
* [bitrise rde session notifications](bitrise_rde_session_notifications.md)	 - List notifications emitted by a session
* [bitrise rde session open-vnc](bitrise_rde_session_open-vnc.md)	 - Open a session's VNC endpoint in the OS-default viewer
* [bitrise rde session restore](bitrise_rde_session_restore.md)	 - Restore a terminated session (re-provisions its VM from the persistent disk)
* [bitrise rde session terminate](bitrise_rde_session_terminate.md)	 - Terminate a running session (preserves it for later restart)
* [bitrise rde session update](bitrise_rde_session_update.md)	 - Update a session's name, description, auto-terminate duration, or labels
* [bitrise rde session upload](bitrise_rde_session_upload.md)	 - Upload a local file or directory into a session
* [bitrise rde session view](bitrise_rde_session_view.md)	 - Show details of a single session
* [bitrise rde session vnc](bitrise_rde_session_vnc.md)	 - Print VNC connection details, or forward the endpoint to a local port

