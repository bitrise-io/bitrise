## bitrise yml

Work with bitrise.yml files.

```
bitrise yml [flags]
```

### Options

```
  -h, --help   help for yml
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

* [bitrise](bitrise.md)	 - Bitrise Automations Workflow Runner
* [bitrise yml get](bitrise_yml_get.md)	 - Print the bitrise.yml stored on Bitrise
* [bitrise yml merge](bitrise_yml_merge.md)	 - Resolves includes in a modular bitrise.yml and merges included config modules into a single bitrise.yml file.
* [bitrise yml update](bitrise_yml_update.md)	 - Upload a new bitrise.yml to Bitrise
* [bitrise yml validate](bitrise_yml_validate.md)	 - Validates a specified bitrise config.

