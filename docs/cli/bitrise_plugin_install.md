## bitrise plugin install

Install bitrise plugin.

```
bitrise plugin install <plugin_source_remote_or_local_url> [flags]
```

### Examples

```
  bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-init.git
  bitrise plugin install ./local-plugin-dir
  bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-init.git --version 1.2.3
```

### Options

```
  -h, --help             help for install
      --version string   Plugin version tag.
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

* [bitrise plugin](bitrise_plugin.md)	 - Plugin handling.

