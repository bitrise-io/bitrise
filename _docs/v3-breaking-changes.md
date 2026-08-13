---
title: Bitrise CLI v3 breaking changes
---

# Bitrise CLI v3 — breaking changes

This document tracks the user-visible breaking changes introduced for the **v3**
major release. v3 merges the cloud resource-management commands into the existing
CLI.

Append new breaking changes here as later v3 steps land. List each change with
what changed, the impact, and how to migrate.

## Legacy (urfave → cobra) cleanup

The v2 line migrated the CLI from `urfave/cli` to `cobra` while preserving the old
surface behind compatibility shims. v3 removes those shims and adopts cobra's
native behavior.

### Argument parsing

- **Single-dash long flags are no longer accepted.** `urfave` treated `-config`
  and `--config` as equivalent; cobra/pflag treat a single dash as shorthand flags.
  So `-workflow` is now rejected (`unknown shorthand flag: 'w' in -workflow`), and —
  worse — `-config x` is silently parsed as the `-c` shorthand with value `onfig`
  (not `--config x`). Always use the double-dash form for long flags:
  `bitrise run -config bitrise.yml` → `bitrise run --config bitrise.yml`. Short flags
  (e.g. `-c`, `-i`) are unaffected.
  *Migrate:* update scripts/CI invocations to use `--<flag>` for long flag names.
- **Unknown flags are now rejected.** Previously an unrecognized flag that followed
  a positional argument was silently ignored (e.g. `bitrise run wf --bogus` still
  ran the workflow). It now produces an error.
- **Unknown commands now produce a concise error.** `bitrise notacommand` prints
  cobra's `unknown command "notacommand" for "bitrise"` error (and exits 1) instead
  of printing the full help text.

### Help and version output

- **Root `--help` uses cobra's native layout.** The previous urfave-style
  `NAME / USAGE / VERSION / GLOBAL OPTIONS / COMMANDS / PLUGINS` layout is gone.
  Installed plugins are still listed (in a `Plugins:` section appended to the help),
  but the `[$ENV]` env-binding hints next to global flags are no longer shown.

### Command listing and completion

- **Commands and flags are listed alphabetically** in help output (previously in
  declaration order).
- **A `completion` command is now available** (cobra's shell-completion generator),
  e.g. `bitrise completion bash`.

### Environment variable handling

Env-var reading for the bool "mode" flags was unified into one consistent rule:
**explicit flag > bound env var (parsed with `strconv.ParseBool`) > inventory-based
default**. A non-bool env value is now an error.

- **`run --secret-filtering` is now bound to `$BITRISE_SECRET_FILTERING`** and
  validated like `trigger --secret-filtering`: the env value is parsed with
  `ParseBool`, a non-bool value errors, and the flag is reported to analytics when
  sourced from the env. Previously `run` matched the env literally (`"true"`/`"false"`
  only), ignored other values, and never reported it as set from the env.
- **`$BITRISE_SECRET_ENVS_FILTERING` is now parsed with `ParseBool`** (e.g. `1`/`0`
  are now accepted) and a non-bool value errors, instead of being matched literally.
- **`$CI` and `$DEBUG` parsing accepts all `ParseBool` spellings** (e.g. `DEBUG=1`
  now enables debug mode). Non-bool values for these already errored and still do.
- An empty value for any of these env vars is treated as unset (the CLI falls back
  to the inventory-based default) rather than as `false`.

## Command reorganization

The commands were regrouped by use case under `local`, `yml`, and `step` parent
commands. The old top-level names continue to work as hidden aliases, so existing
scripts keep running: `trigger-check` was the only command removed outright, and
`validate` the only one whose behavior changed.

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
  command says so on stderr and names the escape hatch, and the result gains a `source`
  field under `--format json`; a local result carries no marker and its output on
  stdout is unchanged from v2. Two flags control it, and they cannot be combined:
  the new `--offline` forces the local-only check, and the optional `--app` (or
  `BITRISE_APP_ID`) enables app-specific checks (stacks, machine types, license
  pools).
  *Migrate:* pass `--offline` anywhere you depend on local validation or on its exact
  output — in particular scripts and tests that assert on validation messages.

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
- **`api_base_url` and `web_base_url` never come from the per-directory file.** These
  two keys — new in v3, managed with `bitrise config get/set/unset/list` and stored in
  the global `~/.config/bitrise/cli/config.yml` — are the one exception to the
  precedence above: they are only ever read from the global file. Both carry
  credentials (a bearer token, a login password) to whatever host they name, and
  `.bitrise-cli.yml` is picked up from the working directory and its ancestors with no
  confirmation, so a repo you merely clone and run `bitrise` inside of must not be able
  to redirect them. `web_base_url` is additionally overridable via
  `$BITRISE_WEB_BASE_URL`, which wins over the global file. *Migrate:* set these keys
  with `bitrise config set` or the env var — either key in a `.bitrise-cli.yml` is
  ignored.
- **`setup`/CLI-update-check/plugin-update-check now also write `config.yml`.** If you
  already have `~/.bitrise/config.json`, it keeps being updated exactly as before (still
  authoritative for reads), and `~/.config/bitrise/cli/config.yml` is kept in sync
  alongside it. If you don't have a legacy file, one is no longer created — these
  commands now write only the new `config.yml`. *Migrate:* no action required.
- **A missing/unwritable `~/.config` directory can now fail `setup` and every
  plugin invocation.** If `~/.config/bitrise/cli/config.yml` can't be written
  (e.g. a CI container where `~/.bitrise` is writable but `~/.config` isn't),
  `bitrise setup` reports failure even though setup itself succeeded, and any
  `bitrise <plugin>` invocation aborts before running the plugin. *Migrate:*
  ensure `~/.config` (or `$XDG_CONFIG_HOME`) is writable wherever `setup` or
  plugins run.
