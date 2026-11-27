## bitrise app quarantined-tests

List and manage the tests quarantined on an app

### Synopsis

List and manage the tests quarantined on an app. Quarantined tests are
excluded from the test runs of the builds the quarantine applies to, so a flaky
test does not fail them.

Changing quarantined tests requires permission to run builds on the app and a
plan that includes test quarantine.

```
bitrise app quarantined-tests [flags]
```

### Options

```
  -h, --help   help for quarantined-tests
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

* [bitrise app](bitrise_app.md)	 - List, inspect, and manage apps.
* [bitrise app quarantined-tests add](bitrise_app_quarantined-tests_add.md)	 - Add a test to quarantine
* [bitrise app quarantined-tests list](bitrise_app_quarantined-tests_list.md)	 - List the tests quarantined on an app
* [bitrise app quarantined-tests remove](bitrise_app_quarantined-tests_remove.md)	 - Remove a test from quarantine
* [bitrise app quarantined-tests update](bitrise_app_quarantined-tests_update.md)	 - Change the comment or scope of a quarantined test
* [bitrise app quarantined-tests view](bitrise_app_quarantined-tests_view.md)	 - Show a quarantined test

