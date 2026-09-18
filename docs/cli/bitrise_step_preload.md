## bitrise step preload

Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps.

### Synopsis

Downloads and caches step versions from the steplib so later runs can use them
without network access. Use the BITRISE_OFFLINE_MODE env var to test after preloading steps.

--minors-since and --patches-since add versions purely by recency, regardless
of --majors/--minors: --minors-since additionally includes the latest patch of
any minor released in the last N months even if its major wasn't selected;
--patches-since additionally includes any patch at all released in the last N
months.

```
bitrise step preload [flags]
```

### Examples

```
  bitrise step preload
  bitrise step preload --maintainer bitrise
  bitrise step preload --majors 3 --minors 2
```

### Options

```
  -h, --help                 help for preload
      --maintainer string    Maintainer of the steps to list or preload (default "bitrise")
      --majors uint          Include X latest major versions (default 2)
      --minors uint          Include X latest minor versions for each major version (default 1)
      --minors-since uint    Include latest patch version of minors that were released in the last X months (default 2)
      --patches-since uint   Include all patch version that were released in the last X months (default 1)
      --steplib-url string   URL of the steplib to list or preload steps from (default "https://github.com/bitrise-io/bitrise-steplib.git")
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

* [bitrise step](bitrise_step.md)	 - Manage steps.

