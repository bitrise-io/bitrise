## bitrise rde session download

Download a file or directory from a session

### Synopsis

Download a file or directory from a running session to the local machine.

The remote path is tar+gzip'd server-side, served via a signed download URL,
then extracted into LOCAL_PATH locally.

When REMOTE_PATH is a directory, the directory itself is recreated inside
LOCAL_PATH by default. Pass --only-contents to drop just its contents into
LOCAL_PATH instead.

```
bitrise rde session download SESSION_ID REMOTE_PATH LOCAL_PATH [flags]
```

### Examples

```
  bitrise rde session download SESSION_ID /Users/vagrant/project/build ./build
  bitrise rde session download SESSION_ID /Users/vagrant/logs ./logs --only-contents
```

### Options

```
  -h, --help            help for download
      --only-contents   when REMOTE_PATH is a directory, extract only its contents (not the directory itself)
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

