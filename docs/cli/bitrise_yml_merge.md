## bitrise yml merge

Resolves includes in a modular bitrise.yml and merges included config modules into a single bitrise.yml file.

### Synopsis

Resolves includes in a modular bitrise.yml and merges included config modules into a single bitrise.yml file.

By default, the command looks for a bitrise.yml in the current directory; a
custom path can be given as a positional argument (unlike 'yml get'/'update'/
'validate', which take it via -c/--config).

With no --output, the merged config and config tree are printed to stdout;
with --output, they're written to bitrise.yml and config_tree.json in that
directory instead.

```
bitrise yml merge [flags]
```

### Examples

```
  bitrise yml merge
  bitrise yml merge ./ci/bitrise.yml
  bitrise yml merge --output ./merged
```

### Options

```
  -h, --help            help for merge
  -o, --output string   Output directory for the merged config file (bitrise.yml) and related config file tree (config_tree.json).
```

### Options inherited from parent commands

```
      --ci             If true it indicates that we're used by another tool so don't require any user input!
      --debug          If true it enables DEBUG mode.
      --no-color       Disable ANSI colors (the NO_COLOR env var is also honored).
      --pr             If true bitrise runs in pull request mode.
  -q, --quiet          Suppress non-error diagnostic messages.
      --theme string   Color theme. Accepted: auto, dark, light, none.
```

### SEE ALSO

* [bitrise yml](bitrise_yml.md)	 - Work with bitrise.yml files.

