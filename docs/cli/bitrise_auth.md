## bitrise auth

Manage the Bitrise access token

### Synopsis

Manage the Bitrise access token used for API requests.

Both Personal Access Tokens (PAT) and Workspace API Tokens (WAT) work the
same way on the wire — paste either kind here.

Storage:
  YAML file at $XDG_CONFIG_HOME/bitrise/cli/auth.yaml (or
  ~/.config/bitrise/cli/auth.yaml). Written with 0600 permissions, separate
  from preferences in config.yml.

Env override:
  BITRISE_TOKEN takes precedence over the saved token; useful for CI.

```
bitrise auth [flags]
```

### Examples

```
  bitrise auth status
  bitrise auth login
  bitrise auth logout
```

### Options

```
  -h, --help   help for auth
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
* [bitrise auth login](bitrise_auth_login.md)	 - Save a Bitrise access token
* [bitrise auth logout](bitrise_auth_logout.md)	 - Remove the saved access token
* [bitrise auth status](bitrise_auth_status.md)	 - Show whether an access token is configured and where it came from

