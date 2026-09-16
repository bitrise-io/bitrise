## bitrise rde device-guide

Print the guide for driving a session's iOS simulator / Android emulator

### Synopsis

Print the device session guide: how to create a session that boots a virtual
device ('rde session create --device-platform ios|android'), wait for the
device to be ready ('rde session view'), connect, drive it efficiently
(accessibility tree first, then input), let a human watch, and what never to
do. Pass ios or android for that platform's specifics.

The guide is Markdown prose; --format json/yml is rejected (there is no
single-object shape for it).

```
bitrise rde device-guide [ios|android] [flags]
```

### Examples

```
  bitrise rde device-guide
  bitrise rde device-guide ios
  bitrise rde device-guide android
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for device-guide
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

