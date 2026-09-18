## bitrise rde session vnc

Print VNC connection details, or forward the endpoint to a local port

### Synopsis

Print the VNC connection details (address, host, port, username, password,
and a ready-to-use vnc:// URL) for a session.

The VNC password is ephemeral and tied to this session. Avoid pasting the
output into chat or sharing it — anyone with the URL can connect to the
session. `rde session view` and other commands intentionally hide it.

In raw mode the URL is the only thing on stdout, so it's safe to pipe:

  open "$(bitrise rde session vnc SESSION_ID)"

In --format json/yml mode a fully-decomposed {address, host, port, username,
password, url} object is emitted — host and port are always discrete fields,
so a caller building its own connection never has to parse the address or URL.

Pass --forward PORT to open an SSH tunnel and expose the session's VNC endpoint
on a local port, then block until Ctrl-C (use 0 to auto-pick a free port):

  bitrise rde session vnc SESSION_ID --forward 0      # auto-pick a local port
  bitrise rde session vnc SESSION_ID --forward 5901   # bind localhost:5901

A native VNC client (macOS Screen Sharing, Remmina, …) can then connect to the
printed localhost address. The tunnel rides the same SSH connection the CLI
already uses, so no direct network route to the session is required and no
credentials are embedded in a URL handed to the OS. Prefer `rde session open-vnc`
when you just want to launch your viewer against a directly-reachable endpoint.

```
bitrise rde session vnc SESSION_ID [flags]
```

### Examples

```
  bitrise rde session vnc SESSION_ID
  bitrise rde session vnc SESSION_ID --format json
  bitrise rde session vnc SESSION_ID --forward 5901
  open "$(bitrise rde session vnc SESSION_ID)"
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
      --forward int     forward the session's VNC endpoint to this local port, then block until Ctrl-C; use 0 to auto-pick a free port
  -h, --help            help for vnc
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

