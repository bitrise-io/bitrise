## bitrise app quarantined-tests update

Change the comment or scope of a quarantined test

### Synopsis

Change the comment or scope of a quarantined test. Only what you pass changes:
--branch replaces the branches but keeps the targets and the comment.

--all-branches and --all-targets remove the branch or target limit, so the
quarantine applies to every branch or every Workflow and Pipeline.

```
bitrise app quarantined-tests update ID [flags]
```

### Examples

```
  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --comment "Tracked in JIRA-123"
  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --branch main --branch "release/*"
  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --all-branches --all-targets
```

### Options

```
      --all-branches     quarantine the test on every branch
      --all-targets      quarantine the test in every Workflow and Pipeline
      --app string       app ID (or set BITRISE_APP_ID)
      --branch strings   quarantine the test only on this branch, replacing the current branches (repeated or comma-separated)
      --comment string   why the test is quarantined; an empty value removes the comment
  -f, --format string    Output format. Accepted: raw (default), json, yml
  -h, --help             help for update
      --target strings   quarantine the test only in this Workflow or Pipeline, replacing the current targets (repeated or comma-separated)
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

