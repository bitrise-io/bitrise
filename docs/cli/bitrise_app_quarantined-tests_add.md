## bitrise app quarantined-tests add

Add a test to quarantine

### Synopsis

Quarantine a test, so it is excluded from the test runs of the app's builds.

A quarantine matches a test case by its exact suites, name and class name, or
by its identifier when the test report gives one. With --build, the test case
is looked up in that build's test results and quarantined with those exact
values, the same way the build page does it. Pass --name as
'bitrise build test-results BUILD_ID' lists it; if more than one test case
matches, --class-name, --suite and --identifier narrow it down.

Without --build, the values are sent as given: --suite takes the name of each
suite the test case belongs to, outermost first, as the test results list
them, which can differ from where the test case is nested.

The quarantine applies to every branch and every Workflow and Pipeline unless
--branch or --target narrows it. Both accept * wildcards.

```
bitrise app quarantined-tests add [flags]
```

### Examples

```
  bitrise app quarantined-tests add --app my-app-id --build abc123 --name LoginTests.testWrongPassword
  bitrise app quarantined-tests add --app my-app-id --build abc123 --name testWrongPassword --suite LoginTests --comment "Flaky on CI"
  bitrise app quarantined-tests add --app my-app-id --build abc123 --name testWrongPassword --branch main --branch "release/*" --target primary
  bitrise app quarantined-tests add --app my-app-id --suite AppTests --suite LoginTests --class-name LoginTests --name testWrongPassword
  bitrise app quarantined-tests add --app my-app-id --suite BullsEyeTests --name testScore --identifier "BullsEyeTests/testScore()"
```

### Options

```
      --app string          app ID (or set BITRISE_APP_ID)
      --branch strings      quarantine the test only on this branch (repeated or comma-separated)
      --build string        ID of a build the test ran in: the test case is looked up in its test results, and the build is recorded as where it was quarantined from
      --class-name string   class name of the test case. With --build, it only narrows the lookup
      --comment string      why the test is quarantined
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for add
      --identifier string   identifier of the test case, if the test report gives one. With --build, it only narrows the lookup
      --name string         name of the test case; with --build, also as Class.name
      --suite stringArray   name of a suite the test case belongs to; repeat it for each, outermost first. With --build, it only narrows the lookup
      --target strings      quarantine the test only in this Workflow or Pipeline (repeated or comma-separated)
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

