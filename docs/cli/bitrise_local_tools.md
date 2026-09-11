## bitrise local tools

Manage available tools from inside the workflow.

```
bitrise local tools [flags]
```

### Options

```
  -h, --help   help for tools
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
* [bitrise local tools catalog](bitrise_local_tools_catalog.md)	 - List officially supported tools
* [bitrise local tools info](bitrise_local_tools_info.md)	 - Show information about installed or active tools.
* [bitrise local tools install](bitrise_local_tools_install.md)	 - Install a specific tool version
* [bitrise local tools latest](bitrise_local_tools_latest.md)	 - Query the latest version of a tool
* [bitrise local tools setup](bitrise_local_tools_setup.md)	 - Install tools from version files or bitrise.yml
* [bitrise local tools versions](bitrise_local_tools_versions.md)	 - List available versions for a supported tool

