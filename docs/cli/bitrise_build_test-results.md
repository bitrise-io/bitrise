## bitrise build test-results

List a build's test suites and test cases with their results

### Synopsis

List the test suites a build reported and the test cases in them, with each
case's status and duration, and the message of failed ones.

Each test case also lists the test suites it belongs to, which can differ from
where it is nested, and the identifier its test report gives it, if any.

Test results are available once the build has finished, and only if the plan
of the app's workspace includes test reports. Filters narrow the test cases
listed, and suites left without any are left out; a suite's counts always
cover all of its test cases. --sort orders the test cases within each suite,
and --limit keeps the first ones across suites.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").

```
bitrise build test-results BUILD_ID [flags]
```

### Examples

```
  bitrise build test-results abc123 --app my-app-id
  bitrise build test-results abc123 --app my-app-id --status failed,error
  bitrise build test-results abc123 --app my-app-id --status failed,error --sort duration --order desc --limit 20
  bitrise build test-results abc123 --app my-app-id --flaky
  bitrise build test-results abc123 --app my-app-id --suite LoginTests --name signin
  bitrise build test-results abc123 --app my-app-id --format json
```

### Options

```
      --app string          app ID (or set BITRISE_APP_ID)
      --class-name string   only test cases with this exact class name
      --flaky               only flaky test cases, or with --flaky=false only the ones that are not flaky
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for test-results
      --limit int           list at most this many test cases, up to 1000 (default: all)
      --name string         only test cases whose name, as listed (Class.name), contains this text (case-insensitive)
      --order string        direction of --sort: asc, desc (default asc)
      --sort string         sort the test cases within each suite by: duration, name, status. status puts error first, then failed, skipped and passed
      --status strings      only test cases with one of these statuses (comma-separated or repeated): passed, failed, skipped, error
      --suite string        only the test suite with this exact name, at any depth, with the suites nested in it
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

* [bitrise build](bitrise_build.md)	 - Trigger, list, and inspect builds.

