## bitrise local run

Runs a specified Workflow.

### Synopsis

Run a workflow defined in bitrise.yml on the local host.

WORKFLOW_ID is normally given as a positional argument. --workflow is also
accepted and takes precedence over the positional argument if both are given.

By default, bitrise.yml is read from the current directory; use --config to
point elsewhere.

```
bitrise local run WORKFLOW_ID [flags]
```

### Examples

```
  bitrise run primary
  bitrise run primary --config ./ci/bitrise.yml
  bitrise run primary --inventory .bitrise.secrets.yml
  bitrise run primary --workflow primary
```

### Options

```
  -c, --config string               Path where the workflow config file is located.
      --config-base64 string        base64 encoded config data.
  -h, --help                        help for run
  -i, --inventory string            Path of the inventory file.
      --inventory-base64 string     base64 encoded inventory data.
      --json-params string          Specify command flags with json string-string hash.
      --json-params-base64 string   Specify command flags with base64 encoded json string-string hash.
      --output-format string        Log format. Available values: json, console
      --secret-filtering            Hide secret values from the log.
      --workflow string             workflow id to run (takes precedence over the positional WORKFLOW_ID argument)
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

* [bitrise local](bitrise_local.md)	 - Run and manage Bitrise workflows on the local host.

