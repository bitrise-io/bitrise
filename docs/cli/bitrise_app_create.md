## bitrise app create

Register a new app on Bitrise

### Synopsis

Register a new app on Bitrise.

--workspace is only required if you belong to several workspaces and haven't
set a default (see 'bitrise config set default_workspace_id').

bitrise.yml handling:
  --bitrise-yml PATH                upload that file as the app's config
  (no flag, ./bitrise.yml exists)   upload it
  (no flag, no file)                skip — server preset for --project-type applies

The new app's ID is saved as the default app_id in
~/.config/bitrise/cli/config.yml, so later commands (e.g. 'bitrise yml get')
target it without --app.

```
bitrise app create [flags]
```

### Examples

```
  bitrise app create
  bitrise app create --repo-url https://github.com/me/proj --workspace acme
  bitrise app create --bitrise-yml ./ci/bitrise.yml --stack osx-xcode-16.0.x
  bitrise app create --format json
```

### Options

```
      --bitrise-yml string    path to bitrise.yml to upload (default: ./bitrise.yml if present, else skip)
      --branch string         default branch (default: 'git symbolic-ref --short HEAD', else "main")
  -f, --format string         Output format. Accepted: raw (default), json, yml
  -h, --help                  help for create
      --project-type string   project type for server-side preset (default "other")
      --provider string       git provider: auto (registers as 'custom' — does not detect the host from --repo-url), github, gitlab, bitbucket, custom (default "auto")
      --public                create as a public app
      --repo-url string       git repo URL (default: 'git remote get-url origin' in cwd)
      --stack string          build stack ID (default "ubuntu-resolute-26.04-bitrise-2026-android")
      --title string          app title (default: last path segment of repo URL)
      --workspace string      workspace ID to own the app (or set BITRISE_WORKSPACE_ID / default_workspace_id; auto-detected if you have exactly one)
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

* [bitrise app](bitrise_app.md)	 - List, inspect, and manage apps.

