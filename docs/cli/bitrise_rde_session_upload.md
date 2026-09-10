## bitrise rde session upload

Upload a local file or directory into a session

### Synopsis

Upload a local file or directory into a running session.

The local path is tarred + gzipped, uploaded to cloud storage via a signed
URL, then extracted on the session VM at REMOTE_FOLDER.

For directories: the directory's contents are extracted into REMOTE_FOLDER
(not the directory itself).

```
bitrise rde session upload SESSION_ID LOCAL_PATH REMOTE_FOLDER [flags]
```

### Examples

```
  bitrise rde session upload SESSION_ID ./project /Users/vagrant/project
  bitrise rde session upload SESSION_ID ./build.tar.gz /Users/vagrant/artifacts
```

### Options

```
  -h, --help   help for upload
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

