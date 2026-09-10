---
title: Bitrise CLI v3 breaking changes
---

# Bitrise CLI v3 — breaking changes

This document tracks the user-visible breaking changes introduced for the **v3**
major release. v3 merges the cloud resource-management commands into the existing
CLI.

**Audience: existing Bitrise CLI v2 users.** Every entry describes a change against
v2 behavior. Commands that are new in v3 are not covered here.

## Command-line behavior

### Argument parsing

- **Single-dash long flags are no longer accepted.** v2 treated `-config` and
  `--config` as the same flag. A single dash now introduces short flags only, so
  `-config x` would otherwise be read as `-c` with the value `onfig`. Rather than
  misread it, the CLI rejects any single-dash spelling of a long flag:
  `bitrise run -config bitrise.yml` →
  `Error: unknown flag: -config (did you mean --config?)`, exit 1. Short flags
  (e.g. `-c`, `-i`) are unaffected.
  *Migrate:* update scripts/CI invocations to use `--<flag>` for long flag names.
- **Unknown flags are now rejected.** Previously an unrecognized flag that followed
  a positional argument was silently ignored (e.g. `bitrise run wf --bogus` still
  ran the workflow). It now produces an error.
- **Unknown commands now produce a concise error.** `bitrise notacommand` prints
  `unknown command "notacommand" for "bitrise"` and exits 1, instead of printing
  the full help text.

### Error output

- **Errors now go to stderr, not stdout.** In v2 every fatal error was printed to
  stdout, mixed into whatever the command was producing. v3 writes them to stderr
  as `Error: <message>`. The exit code is unchanged (1).
  *Migrate:* capture stderr wherever a script reads the failure reason; a
  stdout-only capture now comes back empty on failure.
- **Under `run --output-format json`, fatal errors are no longer JSON.** They used to
  be printed to stdout as a JSON log line like every other message; they are now the
  same plain `Error: …` line on stderr. This covers errors that abort the command —
  bad arguments, no workflow specified, setup failures. A workflow that runs and
  *fails* is unaffected: its log output is still JSON on stdout, and it still exits
  with the build's exit code.
  *Migrate:* read stderr for the abort reason, or key off the exit code.
- **A bare `bitrise` no longer prints an empty `Error:` line** below its help output.

### Help and version output

- **`bitrise --help` has a new layout.** The previous
  `NAME / USAGE / VERSION / GLOBAL OPTIONS / COMMANDS / PLUGINS` layout is gone.
  Installed plugins are still listed, in a `Plugins:` section appended to the help,
  but the `[$ENV]` hints next to global flags are no longer shown.

### Command listing and completion

- **Commands and flags are listed alphabetically** in help output. v2 listed them
  in a fixed, hand-picked order.
- **A `completion` command is now available** for generating shell-completion
  scripts, e.g. `bitrise completion bash`.

### Environment variable handling

Reading the boolean "mode" flags from environment variables was unified into one
rule: **explicit flag > environment variable > inventory-based default**. The
accepted values are `1`, `t`, `T`, `true`, `TRUE`, `True` and their false
counterparts (`0`, `f`, `F`, `false`, `FALSE`, `False`); anything else is now an
error.

- **`run --secret-filtering` now reads `$BITRISE_SECRET_FILTERING`**, the way
  `trigger --secret-filtering` always did: any accepted spelling works, a
  non-boolean value errors, and the flag is reported to analytics when it came from
  the environment. Previously `run` matched the value literally (`"true"`/`"false"`
  only), ignored anything else, and never reported it as set from the environment.
- **`$BITRISE_SECRET_ENVS_FILTERING` accepts every boolean spelling** (e.g. `1`/`0`
  now work) and errors on a non-boolean value, instead of being matched literally.
- **`$CI` and `$DEBUG` accept every boolean spelling** (e.g. `DEBUG=1` now enables
  debug mode). Non-boolean values for these already errored and still do.
- An empty value for any of these variables is treated as unset — the CLI falls back
  to the inventory-based default rather than to `false`.

## Command reorganization

The commands were regrouped by use case under `local`, `yml`, and `step` parent
commands. The old top-level names continue to work as hidden aliases, so existing
scripts keep running; `trigger-check` is the only command removed outright.

### `trigger-check` removed

- **`bitrise trigger-check` no longer exists.** It had not been updated with newer
  trigger features for a long time and was unused.
  *Migrate:* remove `bitrise trigger-check` invocations from scripts. There is no
  direct replacement; `bitrise trigger` still runs a workflow by trigger params.

### `trigger` hidden

- **`bitrise trigger` is now hidden** from help output. It still works for backward
  compatibility but is deprecated.

### Commands grouped under `local`, `yml`, and `step`

- **The canonical command paths changed.** Each command now lives under a parent
  that reflects its use case. The old top-level names are kept as hidden aliases,
  so they keep working, but help and documentation refer to the new paths.

  | Old (still works) | New canonical path |
  | --- | --- |
  | `bitrise run` | `bitrise local run` |
  | `bitrise init` | `bitrise local init` |
  | `bitrise setup` | `bitrise local setup` |
  | `bitrise tools …` | `bitrise local tools …` |
  | `bitrise workflows` | `bitrise local workflows` |
  | `bitrise validate` | `bitrise yml validate` |
  | `bitrise merge` | `bitrise yml merge` |
  | `bitrise steps list-cached` | `bitrise step list-cached` |
  | `bitrise steps preload` | `bitrise step preload` |
  | `bitrise share …` | `bitrise step share …` |

  *Migrate:* no action required for existing scripts. New usage and documentation
  should prefer the grouped paths.

### Plugins named like a new command need the colon prefix

- **A plugin whose name collides with a command is no longer reachable without `:`.**
  A plugin is only looked up when the first word isn't a known command, so the
  commands added in v3 now shadow same-named plugins: `bitrise step …` runs the built-in
  `step` command instead of the bundled `:step` plugin. The same applies to any plugin
  named `app`, `build`, `auth`, `stack`, `user`, `config`, `api`, `rde`, `yml` or
  `local`.
  *Migrate:* use the colon syntax — `bitrise :step …` — which has always worked and is
  unaffected.

### `yml validate` updated

- **`validate` now validates online when you're authenticated, instead of locally.**
  When an access token resolves (from the new `bitrise auth login` command or
  `$BITRISE_TOKEN`), the config is submitted to Bitrise and the local schema check is
  skipped entirely — the API runs the same schema checks plus app-specific ones. With
  no token, or when the online attempt can't be completed (network, 5xx, timeout),
  validation falls back to the local check and reports the reason as a warning. Because
  the old top-level `bitrise validate` is an alias of this command, existing invocations
  change behavior as soon as a token is present — including picking up any difference
  between the server's messages and the local ones. When validation happens online the
  command says so on stderr and tells you how to force the local check, and the
  result gains a `source`
  field under `--format json`; a local result carries no marker and its output on
  stdout is unchanged from v2. Two flags control it, and they cannot be combined:
  the new `--offline` forces the local-only check, and the optional `--app` (or
  `BITRISE_APP_ID`) enables app-specific checks (stacks, machine types, license
  pools).
  *Migrate:* pass `--offline` anywhere you depend on local validation or on its exact
  output — in particular scripts and tests that assert on validation messages.
- **An unsupported `format_version` no longer fails validation online.** v2 rejected a
  `bitrise.yml` whose `format_version` is newer than the CLI supports — `is_valid:
  false`, a hard error, exit 1. On the online path that check is now only a warning,
  and `is_valid` follows whatever the API reports, so the same file validates clean
  and exits 0. Since a build normally has a token, this is the common case in CI.
  `--offline` still fails it exactly as v2 did.
  *Migrate:* pass `--offline` wherever validation is meant to gate on the running CLI
  being new enough for the config.
- **Inside a Bitrise build, `bitrise validate` now runs app-specific checks.**
  `$BITRISE_APP_SLUG` is injected into every build and is now read as an app-ID
  fallback, so an authenticated `bitrise validate` with no app named validates against
  the app the build runs for. That adds the API's app-specific checks — stacks, machine
  types, license pools — which previously ran for nobody, so a config that validated
  cleanly before can now report app-specific errors. `--offline` is unaffected: it
  skips the online path, and an app ID picked up from the environment stays ignored
  there.
  *Migrate:* pass `--offline`, or unset `BITRISE_APP_SLUG` for the invocation, to
  restore the previous behavior.
- **`validate --format yml` now works.** v2's help advertised "Accepted: json, yml"
  but the command rejected anything other than `raw` and `json`, exiting 1 with
  `Invalid format: yml`. It now renders the result as YAML. With no `--format`, the
  command follows the new global `--output` / `$BITRISE_OUTPUT` / `output` config key
  — all unset by default, so nothing changes unless you opt in.

## Config handling

- **Two new config file locations are now read, layered under the existing one.**
  Besides the pre-existing `~/.bitrise/config.json`, the CLI now also reads a global
  `~/.config/bitrise/cli/config.yml` and a per-directory `.bitrise-cli.yml` (found by
  searching the working directory and its ancestors). Precedence, highest to lowest:
  `~/.bitrise/config.json` > `.bitrise-cli.yml` > `~/.config/bitrise/cli/config.yml`.
  The pre-existing legacy file stays authoritative — if it sets a value, that value
  wins over anything in the new files, so nothing changes for users who already have
  one. *Migrate:* no action required. To have a value controlled by the new
  per-directory or global file instead, remove that value from
  `~/.bitrise/config.json`.
- **`setup`/CLI-update-check/plugin-update-check now also write `config.yml`.** If you
  already have `~/.bitrise/config.json`, it keeps being updated exactly as before (still
  authoritative for reads), and `~/.config/bitrise/cli/config.yml` is kept in sync
  alongside it. If you don't have a legacy file, one is no longer created — these
  commands now write only the new `config.yml`. *Migrate:* no action required.
- **An unwritable `~/.config` directory can now fail `setup` and plugin invocations.**
  A *missing* directory is fine — it's created, at any depth. An unwritable one only
  matters if you have no `~/.bitrise/config.json` yet: when that legacy file exists it
  stays authoritative and a failed `config.yml` write is downgraded to a warning.
  Without it, `bitrise setup` reports failure even though setup itself succeeded, and
  a `bitrise <plugin>` invocation aborts before running the plugin — on the runs where
  that plugin's update check is due. *Migrate:* ensure `~/.config` (or
  `$XDG_CONFIG_HOME`) is writable wherever `setup` or plugins run.

## Telemetry

- **`ANALYTICS_DISABLED=true` now disables analytics as well.** v2 honored only
  `BITRISE_ANALYTICS_DISABLED`; both names work now. A side effect is visible on
  stdout: `bitrise run` stops printing its "Bitrise collects anonymous usage stats…"
  notice for anyone who had the unprefixed variable set. *Migrate:* no action
  required; unset `ANALYTICS_DISABLED` if you were relying on Bitrise analytics
  staying on while another tool's were off.
