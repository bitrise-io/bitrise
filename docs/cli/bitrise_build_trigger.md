## bitrise build trigger

Start a new build

### Synopsis

Start a new build on the given app.

The app is resolved via --app ID, BITRISE_APP_ID, or "bitrise config set app_id ID".

If neither --workflow nor --pipeline is given, Bitrise selects the
appropriate workflow from the trigger map.

--branch defaults to "main" unless --tag or --commit-hash is given.

--env values are expanded ($VAR references are replaced) on the build machine.
List keys in --env-no-expand to pass their values verbatim.

--priority overrides the priority set in bitrise.yml and the trigger map,
including an explicit 0. Omit it to keep those.

--wait blocks until the build finishes without streaming logs; with --format
json/yml the final build record is written to stdout.

--watch waits for the build to finish, showing progress: with --format
json/yml, build logs stream to stderr as plain text and the final build
record is written to stdout; otherwise, on a terminal, an interactive status
display is shown instead of raw log lines, and when stdout isn't a terminal,
logs stream as plain text there instead.

```
bitrise build trigger [flags]
```

### Examples

```
  bitrise build trigger --app my-app-id --workflow primary
  bitrise build trigger --app my-app-id --workflow deploy --branch release/1.2 --format json
  bitrise build trigger --app my-app-id --pipeline my-pipeline --branch main
  bitrise build trigger --app my-app-id --workflow primary --tag v1.2.3
  bitrise build trigger --app my-app-id --workflow primary --branch-dest main --pull-request-id 42
  bitrise build trigger --app my-app-id --workflow primary --env '{"MY_VAR":"hello","OTHER":"world"}'
  bitrise build trigger --app my-app-id --workflow primary --env '{"API_URL":"https://example.com","PRICE":"$5"}' --env-no-expand PRICE
  bitrise build trigger --app my-app-id --workflow primary --stack osx-xcode-16.0.x --machine-type g2-m1.4core
  bitrise build trigger --app my-app-id --pipeline my-pipeline --priority 0
  bitrise build trigger --app my-app-id --workflow primary --wait
  bitrise build trigger --app my-app-id --workflow primary --watch
```

### Options

```
      --app string              app ID (or set BITRISE_APP_ID)
      --branch string           branch to build (default "main" unless --tag or --commit-hash is given)
      --branch-dest string      target branch for pull-request builds
      --commit-hash string      commit hash to build
      --commit-message string   commit message to record
      --env string              environment variables as a JSON object, e.g. '{"KEY":"value"}'
      --env-no-expand strings   keys from --env whose values are passed verbatim, without expanding $VAR references (repeatable or comma-separated)
  -f, --format string           Output format. Accepted: raw (default), json, yml
  -h, --help                    help for trigger
      --interval duration       polling interval when --wait or --watch is active (default 3s)
      --license-pool string     license pool ID to run the build with
      --machine-type string     machine type ID to run the build on, overrides the workflow's machine type
      --pipeline string         pipeline ID to trigger (mutually exclusive with --workflow)
      --priority int            build priority from -100 to 100, overrides the bitrise.yml and trigger map priority (available on certain plans only)
      --pull-request-id int     pull request ID for PR builds
      --stack string            stack ID to run the build on, overrides the workflow's stack (see 'bitrise stack list')
      --tag string              tag to build
      --wait                    block until the build finishes without streaming logs (exit code reflects build outcome)
      --watch                   wait for the build to finish, showing progress (exit code reflects build outcome)
      --workflow string         workflow ID to trigger (mutually exclusive with --pipeline)
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

