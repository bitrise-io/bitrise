## bitrise app quarantined-tests remove

Remove a test from quarantine

### Synopsis

Remove a test from quarantine, so it is included in test runs again.

Asks for confirmation unless --yes is passed. Without a terminal to ask on,
or with --ci, --yes is required.

```
bitrise app quarantined-tests remove ID [flags]
```

### Examples

```
  bitrise app quarantined-tests remove 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id
  bitrise app quarantined-tests remove 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --yes
```

### Options

```
      --app string   app ID (or set BITRISE_APP_ID)
  -h, --help         help for remove
      --yes          remove without asking for confirmation
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

* [bitrise app quarantined-tests](bitrise_app_quarantined-tests.md)	 - List and manage the tests quarantined on an app

