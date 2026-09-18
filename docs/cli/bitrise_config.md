## bitrise config

Manage CLI configuration (defaults persisted to a YAML file).

### Synopsis

Manage persistent CLI configuration.

Storage:
  Global file: ~/.config/bitrise/cli/config.yml
               (honors $XDG_CONFIG_HOME instead of ~/.config)
  Per-dir:     .bitrise-cli.yml in the current directory or any ancestor

Recognized keys: api_base_url, web_base_url, rde_api_base_url, app_id, default_workspace_id, output, theme

api_base_url, web_base_url, and rde_api_base_url resolve as: global file >
built-in default. All three are additionally overridable via $BITRISE_API_BASE_URL, $BITRISE_WEB_BASE_URL and
$BITRISE_RDE_API_BASE_URL, which win over the global file. They
deliberately ignore the per-directory .bitrise-cli.yml — each names a host
that receives credentials, and a repo you merely clone and run 'bitrise'
inside of must not be able to silently redirect them.

app_id resolves as: --app flag > $BITRISE_APP_ID > $BITRISE_APP_SLUG >
per-directory file > global file. default_workspace_id resolves the same way,
via --workspace and $BITRISE_WORKSPACE_ID. Both honor the per-directory file
precisely so a repo can pin which app and workspace its checkout belongs to;
they're identifiers, not credentials. 'bitrise app create' writes app_id to the
global file, and falls back to default_workspace_id when --workspace is omitted.

Bitrise sets $BITRISE_APP_SLUG and $BITRISE_WORKSPACE_ID in every build, so
inside a build a command with no --app/--workspace acts on the app the build
runs for and the workspace owning it. Pass the flag explicitly to target
anything else.

output resolves as: --output flag > $BITRISE_OUTPUT > output config key (per-directory
file then global file) > built-in default (raw). theme resolves the same way,
via --theme and $BITRISE_CLI_THEME (default: auto). Both honor the per-directory file, like
app_id/default_workspace_id above — neither is a credential or a URL. Note
--output only affects commands that share the raw/json/yml format vocabulary:
'local workflows' and 'plugin list/info' keep their own --format flag.

'get'/'set'/'unset'/'list' only read and write the global file — per-dir
files must be edited by hand.

To manage your access token, use 'bitrise auth login/logout/status'.

```
bitrise config [flags]
```

### Options

```
  -h, --help   help for config
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
* [bitrise config get](bitrise_config_get.md)	 - Print the value of a single config key
* [bitrise config list](bitrise_config_list.md)	 - List the values currently saved in the global config file
* [bitrise config path](bitrise_config_path.md)	 - Print the absolute path of the global config file
* [bitrise config set](bitrise_config_set.md)	 - Set a config key and save the global config file
* [bitrise config unset](bitrise_config_unset.md)	 - Remove a config key and save the global config file

