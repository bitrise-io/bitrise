## bitrise auth status

Show whether an access token is configured and where it came from

### Synopsis

Show whether an access token is configured and which source supplied it.

Sources, in precedence order:
  env        BITRISE_TOKEN environment variable
  auth file  auth.yaml, written by 'bitrise auth login' (OAuth or a
             pasted/email token — a new login overwrites the previous one).
             OAuth logins are shown as "oauth (auth file)" and refreshed
             automatically.
  none       no token configured

```
bitrise auth status [flags]
```

### Options

```
  -f, --format string   Output format. Accepted: raw (default), json, yml
  -h, --help            help for status
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

* [bitrise auth](bitrise_auth.md)	 - Manage the Bitrise access token

