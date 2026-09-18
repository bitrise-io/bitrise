## bitrise plugin

Plugin handling.

```
bitrise plugin [flags]
```

### Options

```
  -h, --help   help for plugin
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
* [bitrise plugin delete](bitrise_plugin_delete.md)	 - Delete bitrise plugin.
* [bitrise plugin info](bitrise_plugin_info.md)	 - Installed bitrise plugin's info
* [bitrise plugin install](bitrise_plugin_install.md)	 - Install bitrise plugin.
* [bitrise plugin list](bitrise_plugin_list.md)	 - List installed bitrise plugins.
* [bitrise plugin update](bitrise_plugin_update.md)	 - Update bitrise plugin. If <plugin_name> not specified, every plugin will be updated.

