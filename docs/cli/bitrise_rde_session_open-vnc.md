## bitrise rde session open-vnc

Open a session's VNC endpoint in the OS-default viewer

### Synopsis

Hand the session's VNC URL to the operating system's default URL handler:

  - macOS:    /usr/bin/open
  - Linux:    xdg-open (must be installed; install x11-utils or similar)
  - Windows:  cmd /c start

The OS launches whatever app is registered for vnc:// (Screen Sharing on
macOS by default; Remmina/Vinagre on Linux; a third-party client on Windows).

The URL contains the ephemeral VNC password as a userinfo component. The
URL is passed as an argv element to the OS handler, so it is briefly
visible to other processes on the machine that can read this process's
argv (e.g. `ps`). On a single-user dev machine this is usually fine;
on a shared host, prefer `rde session vnc` and paste the URL into your
viewer manually.

```
bitrise rde session open-vnc SESSION_ID [flags]
```

### Examples

```
  bitrise rde session open-vnc SESSION_ID
  bitrise rde session open-vnc SESSION_ID --format json
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for open-vnc
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

