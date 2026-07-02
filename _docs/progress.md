---
title: CLI merge — progress & design notes
---

# CLI merge — progress & design notes

Internal working notes for merging the two Bitrise CLIs. Kept so we don't have to
re-read the RFC each session. Audience: maintainers working on the merge (not users —
user-facing changes go in `v3-breaking-changes.md`).

## 1. Purpose & links

Merge the cloud resource-management CLI (`bitrise-cli`) into the existing local CLI
(`bitrise`), so existing users get the new features through the current auto-update
channel, and there is one CLI to maintain.

- RFC "Extend Bitrise CLI with resource management commands" — Confluence page
  `5042241595` (space ENGI), status ACCEPTED.
- New CLI repo: `github.com/bitrise-io/bitrise-cli` (checked out at `../bitrise-cli`).
- This repo: `github.com/bitrise-io/bitrise/v2`.
- Jira epic BACKEND-568 "Merge into old CLI" tracks the effort; subtasks BACKEND-569..578.
- Execution is a stack of small, single-purpose PRs on a long-lived `v3-development`
  branch (merged to `master` later, when it makes sense), reviewed in order. Landed so
  far: remove legacy compatibility hacks (breaking changes → `v3-breaking-changes.md`) →
  command reorganization (RFC §2 shape, no functional change) → split commands into
  packages (this doc's §5-§6, no functional change, BACKEND-573).

## 2. RFC target — "Commands after the merge"

Parent (group) commands host both old and new subcommands. Old top-level names are
kept as **hidden aliases** so existing scripts keep working.

- **`local`** — local workflow runner: `local run/init/setup/tools/workflows`
  (aliases: `run`, `init`, `setup`, `tools`, `workflows`).
- **`yml`** — `yml merge` (alias `merge`), `yml validate` (alias `validate`).
  Future online (from new CLI): `yml get`, `yml update`, online `yml validate`.
  RFC: alias old+new `validate` into one command — do online validation when
  authenticated, fall back to local otherwise.
- **`step`** — `step list-cached` + `step preload` (aliases `steps …`), `step share`
  (alias `share`). Future online (from new CLI): `step search`, `step inputs`
  (no overlap with the cache/share subcommands → they just coexist).
- Unchanged top-level: `plugin`, `envman`, `update`, `version`, `help`.
- `trigger` kept but **hidden**; `trigger-check` **removed** (both already done).
- Future new top-level groups (added in the v3 merge): `app`, `auth`, `build`,
  `config`, `user`, `rde`, `stack`, `api`, `purr`, `completion`.
- Global flags to unify later: old `--debug`/`--ci`/`--pr` + new `--output`/`-o`,
  `--quiet`/`-q`, `--no-color`, `--theme`.

## 3. The two codebases

| | old `bitrise` (this repo) | new `bitrise-cli` |
|---|---|---|
| Module | `github.com/bitrise-io/bitrise/v2` | `github.com/bitrise-io/bitrise-cli` |
| Command pkg | flat `cli` (`cli/*.go`) | `cmd` + per-group sub-pkgs `cmd/<group>/` |
| Group pattern | factory funcs in one pkg | each group a pkg exposing `NewCmd()` |
| Entrypoint | `main.go` → `cli.Run()` | `main.go` → `cmd.Execute()` |
| Shared helpers | scattered in `cli` pkg | `cmd/cmdutil/` |
| Service layer | domain pkgs (`bitrise/`, `tools/`, …) | `internal/` (app/build/rde/config/auth/output/…) |
| API client | — | `bitriseapi/` (+ `bitriseapi/rde`) |
| Global flags | `--debug/--ci/--pr` | `--output/-o`, `--quiet/-q`, `--no-color`, `--theme` |
| Config | `~/.bitrise/config.json` (JSON) | `~/.config/bitrise/config.yaml` + `auth.yaml` + per-dir `.bitrise-cli.yml` (XDG, YAML), precedence chain |
| DI | package globals + process env | `config.Resolved` on `cmd.Context()` (`config.WithResolved`/`FromContext`) |
| Output | global JSON logger (`log/`), `--output-format` on `run` | `internal/output` (human/json) + `internal/output/style` (lipgloss themes) |
| Analytics | `analytics/` (tracked per command) | none |
| Self-update | `cli/update.go` | `cmd/update.go` (`notifyUpdateAvailable`/`armUpdateCheck`) |

New CLI group shape (mirror target): `bitrise-cli/cmd/app/cmd.go`, `cmd/rde/cmd.go` —
`package <group>`, `func NewCmd() *cobra.Command`, parent sets flags + `AddCommand(newXCmd())`,
one file per subcommand. Root: `bitrise-cli/cmd/root.go` — `rootCmd.AddCommand(group.NewCmd())`.

The old repo has NO `internal/`/`bitriseapi/`/`cmdutil/` (no collision); it HAS
`analytics/`, `configs/`, `output/`, `log/` (old equivalents) and uses `vendor/`.

## 4. Decisions

1. **Mirror the new structure.** Split flat `cli` into per-group sub-packages under
   `cli/`: `cli/local`, `cli/yml`, `cli/step`, `cli/plugin`, each exposing `NewCmd()`.
2. **Root stays `cli/`.** Root wiring stays in package `cli`; `cli/commands.go` renamed
   to `cli/root.go`. Entrypoint stays `cli.Run()` — `main.go` unchanged.
3. **Shared helpers → `cli/cmdutil`** (package `cmdutil`), mirroring `cmd/cmdutil`. Kept
   lightweight. Leaf package (imports neither root nor groups) → no import cycle.
4. **Runner engine → `cli/local`.** The WorkflowRunner engine (step activation/env,
   build-result collector, agent hooks) is used only by `run` and `trigger`, so it lives
   in `cli/local`. `trigger` also moves to `cli/local` and is **hidden everywhere**: a
   hidden `local trigger` subcommand AND the hidden top-level `trigger` alias, both from
   one `local.NewTriggerCommand()` constructor.
5. **`bitriseapi/` top-level, `internal/` top-level** — created in the *later* v3 merge,
   not now. `cmdutil` stays under `cli/` (the future `cmd/cmdutil` merges into it).
6. **Scope now = structure + docs only.** No bitrise-cli command/service/API code ported.

## 5. Reorg design

### Package / import graph (acyclic)
```
main.go → cli(root)
cli(root) → cli/local, cli/yml, cli/step, cli/plugin, cli/cmdutil
cli/{local,yml,step,plugin} → cli/cmdutil
cli/cmdutil → cli/docker, cli/containermanager, external only   (leaf)
```
Pre-cobra dispatch (`detectPlugin`, `runPlugin`, `envmanPassthrough`, `runEnvman`) stays
in package `cli` (`cli_test.go` pins it there). It calls `cmdutil.CommandTokenIndex` /
`cmdutil.GlobalFlagNames` and iterates `root.Commands()` — works regardless of which
package built each command.

### File-move map
- **Stay root `cli`:** `cli.go`, `commands.go`→`root.go`, `help.go`, `envman.go`,
  `version.go`, `update.go` (command shell; update-check cluster → cmdutil),
  plugin detect/run dispatch (→ `plugin_dispatch.go`).
- **→ `cli/cmdutil`:** `flags.go`, `modes.go`, `command_analytics.go`, `args.go`,
  `json_output.go` + config/inventory + params + update-cluster helpers (exported).
- **→ `cli/local`:** `local.go`(→cmd.go), `run.go`, `init.go`, `setup.go`, `tools.go`,
  `workflow_list.go`, `trigger.go`, and the engine: `run_util.go`, `run_config.go`,
  `run_trigger_params.go`, `build_run_result_collector.go`, `agent.go`,
  `step_activator.go`, `step_environment.go`, `analytics.go`.
- **→ `cli/yml`:** `validate.go`, `merge.go` (+ `cmd.go`).
- **→ `cli/step`:** `preload_steps.go`(→cache.go), `share.go`, `share_*.go` (+ `cmd.go`).
- **→ `cli/plugin`:** `plugin_install/update/delete/info/list.go` (+ `cmd.go`).
- **Tests follow their symbols:** `run_*_test.go`, `tools_test.go`, `trigger_test.go`,
  `modes_test.go`, `agent_test.go`, `analytics_test.go` → their group; `cli_test.go` → root.

### `cli/cmdutil` exported symbols
- Errors/tracker: `Failf`, `SetTracker` (root `Run()` inits the tracker), `LogCommandParameters`,
  `LogPluginCommandParameters`, `SendCommandInfo`, `SetFlagEnvVar`, `EnvVarAnnotation`.
- Flags: all `*Key` consts (incl. `OutputFormatKey`, relocated from run.go), `GlobalFlagNames`,
  `AddConfigAndInventoryFlags`, `AddJSONParamsFlags`, `AddSecretFilteringFlag`, `AddTriggerFilterFlags`.
- Dispatch/help: `CommandTokenIndex`, `ApplyGlobalFlagsFromArgs`, `IsFlag`,
  `RequireKnownSubcommand`, `AsHidden`, `ShowSubcommandHelp`, json loggers.
- Modes: `ResolveBoolEnv`, `ResolveBoolFlagOrEnv`, `IsPRMode`, `IsCIMode`, `IsSecretFiltering`,
  `IsSecretEnvsFiltering`, `IsSteplibOfflineMode`, `RegisterSteplibOfflineMode`.
- Config/inventory (yml + local): `CreateBitriseConfigFromCLIParams`, `CreateInventoryFromCLIParams`,
  `GetBitriseConfigFilePath`, `GetInventoryFilePath`, `GetBitriseConfigFromBase64Data`,
  `GetInventoryFromBase64Data`.
- Params (run+trigger): `RunAndTriggerParamsModel`, `ParseRunParams`, `ParseTriggerParams`,
  `ParseRunAndTriggerJSONParams`, `ParseRunAndTriggerParams`.
- Update cluster (local runner + root `update`): `CheckUpdate`, `LatestTag`, `InstalledWithBrew`,
  `NewVersionFromBrew`, `NewCLIVersion`, `PrintCLIUpdateInfos`.

### Root wiring + alias block
Groups attach via `group.NewCmd()`; old names re-registered as hidden aliases from the
same exported leaf constructors wrapped in `cmdutil.AsHidden` (a `*cobra.Command` binds to
one parent, so each alias is a fresh command from the constructor). `local.NewCmd()` also
mounts `NewTriggerCommand()` hidden (→ hidden `local trigger`).

## 6. yml / step reconciliation

`cli/yml/cmd.go` and `cli/step/cmd.go` `NewCmd()` list existing subcommands with a
commented mount point, so the future merge is "add file + uncomment AddCommand":
```
yml:  AddCommand(NewValidateCommand(), NewMergeCommand())        // + NewGetCommand(), NewUpdateCommand()
step: AddCommand(NewListCachedStepsCommand(), NewPreloadStepsCommand(), NewShareCommand())  // + NewSearchCommand(), NewInputsCommand()
```
`yml validate` alias/fallback: `NewValidateCommand().RunE` runs local validate now; the
merge adds an online branch (auth-detected, `--app`/`--offline` flags) with local fallback —
applies to both `yml validate` and the hidden `validate` alias because both call the one
constructor. Keep `RequireKnownSubcommand` (help on bare) for both parents in v2; do NOT
copy the new CLI's bare-parent default-to-`get`/`list` (that would be a conscious v3 change).

## 7. v3 merge recipe (later tasks)

Rather than redoing config/auth/API/flags/analytics reconciliation inside every
per-group port, those cross-cutting concerns are pulled out as their own upfront
steps. Order below is dependency-driven, checked against actual imports in
`../bitrise-cli` (not just the conceptual grouping) — see reasoning below the list.
May still be split into further sub-PRs as work proceeds:

1. **Port `internal/auth`** (token store only: `Auth` type, `Load`/`Save`/`Clear`/
   `Path`, ~140 lines). Leaf package, zero internal imports. Needed first because
   `config.Resolve` takes an `auth.Auth` param and `cmd/root.go`'s `persistentPreRun`
   calls `auth.Load()` directly — config handling can't compile without it. This is
   **not** the same as the `auth` command (login/logout) — that's `cmd/auth.go` +
   `internal/oauth` (the OAuth device-flow login UX), which depends on this package
   and is ported later as its own command group in step 4.
2. **Extend config handling.** Port/reconcile `internal/config`
   (`../bitrise-cli/internal/config/{config.go,resolve.go}`) into `cli/cmdutil` (or a
   new `cli/config`). Old: `~/.bitrise/config.json`, package-global state, process env
   (§3 DI row). New: `~/.config/bitrise/config.yaml` + `auth.yaml` + per-dir
   `.bitrise-cli.yml`, XDG/YAML, precedence chain, `config.Resolved` threaded via
   `cmd.Context()` (`config.WithResolved`/`FromContext`). Merged `PersistentPreRunE`
   needs to keep old `configs.InitPaths`/mode env AND add new `config.Load`+`Resolve`+
   `WithResolved`+`style.Configure`. Token resolution is "env > auth.yaml"
   (`resolve.go`), so this is fully usable via env-var tokens before the `auth login`
   UX exists.
3. **Port the `bitriseapi` core client only** — `client.go`+`raw.go`+`paging.go`+
   `me.go` (~390 lines: `Client`, `New`, generic `get[T]`/`getPage[T]`/`postDecode[T]`,
   `APIError`, `RawRequest`, paging types). Zero internal imports either way; this is
   shared scaffolding every domain endpoint file builds on. Do **not** port the rest of
   `bitriseapi/` (`apps*.go`, `builds.go`, `organizations.go`, `stacks.go`, `steps.go`,
   `yml.go`, ~1600 more lines) or `bitriseapi/rde/` (~2700 lines) here — nothing
   requires them yet, and porting all 15+ files as one PR is the kind of big-bang dump
   we're trying to avoid. Each domain file rides along with the command group that
   actually needs it (next step).
4. **Add command groups one by one** (`app`, `auth`, `build`, `rde`+nested, `user`,
   `api`, `stack`; plus new files for existing `yml`/`step`, per §6). `auth` (login/
   logout, `cmd/auth.go` + `internal/oauth`) is just one more group here, not special —
   it depends on step 1 (`internal/auth`) same as everything else. Per group `<g>`:
   - Copy the dir: `cp -R ../bitrise-cli/cmd/<g> cli/<g>` (for yml/step, copy only the
     new files).
   - Copy the matching `internal/<g>` service layer (e.g. `internal/app/service.go`)
     and any `bitriseapi/<domain>.go` file(s) it needs that aren't ported yet.
   - Rewrite import prefixes: `…/bitrise-cli/cmd/<g>` → `…/bitrise/v2/cli/<g>`,
     `…/bitrise-cli/cmd/cmdutil` → `…/bitrise/v2/cli/cmdutil`,
     `…/bitrise-cli/bitriseapi` → `…/bitrise/v2/bitriseapi`,
     `…/bitrise-cli/internal/…` → `…/bitrise/v2/internal/…`.
   - Wire root: `rootCmd.AddCommand(<g>.NewCmd())` (or uncomment in yml/step cmd.go).
   - Vendor new deps: `go mod tidy && go mod vendor` (bubbletea, lipgloss, termenv,
     golang.org/x/term). Reconcile Go version (old `go.mod` vs new `go 1.26.2`).
   - `go build ./... && go test ./cli/... && golangci-lint run`.
5. **Extend global flag handling both ways** — reconcile
   `--output/-o/--quiet/--no-color/--theme` vs `--debug/--ci/--pr` (watch `-o`
   collisions); make ported commands read `--output` via `cmdutil.ResolveFormat`.
6. **Make sure analytics work for new commands** — decide whether ported commands call
   `cmdutil.LogCommandParameters` (old CLI has this; new CLI has no analytics today).

Why this order (checked via `grep` imports in `../bitrise-cli`, not assumed): the
original draft had config before auth and a monolithic "API support" step; both were
wrong. `internal/config` imports `internal/auth` (not the reverse) — auth must land
first. `bitriseapi/` (incl. `rde/`) has no internal imports in either direction, so
it's not a hard prerequisite for config/auth at all — it only matters once a command
group needs it, hence splitting it into a small shared core now and per-domain files
later.

Known conflicts to decide (user-visible): config location `~/.bitrise` vs `~/.config/bitrise`;
auth token store (new); output/log (two systems — keep stdout=data / stderr=diagnostics so
JSON stays pipeable); self-update (two mechanisms — pick one); bare-parent default per group.

## 8. Task status

Jira epic **BACKEND-568** "Merge into old CLI" (In Progress) tracks the stack; see
subtasks BACKEND-569..578.

- BACKEND-571 — migrate old commands from urfave to Cobra — **done**, PR approved but
  not yet merged (Mobile team owns the phased rollout).
- BACKEND-573 — reorganize into per-group sub-packages — **done**, committed as
  `dc569f05` ("split commands into packages to prepare for new structure") on
  `BACKEND-573-command-folders`, pushed. (Jira title is "Move new commands into old
  CLI" — actual scope for this PR was structure-only, per decision §4.6.) No
  functional change. Final `cli/` tree:
  `cli/{cli.go,root.go,version.go,update.go,envman.go,help.go,cli_test.go,
  plugin_dispatch.go}` (root, package `cli`) + `cli/{cmdutil,local,yml,step,plugin,
  docker,containermanager}/` sub-packages, matching the design in §5-§6 with no
  deviations; no import cycles. One unrelated pre-existing test failure:
  `TestParseAndValidatePluginFromYML` (`plugins/model_methods_test.go`) fails because
  `stepman` isn't installed in this sandbox — confirmed via `git stash` that it fails
  identically on the untouched tree.
- **BACKEND-573-token-store — port `internal/auth` (§7 step 1) — done**, on branch
  `BACKEND-573-token-store` (stacked on `BACKEND-573-command-folders`), not yet
  committed — left for the user to review/commit. Added `internal/auth/{auth.go,
  auth_test.go}` — this repo's first top-level `internal/` package — ported from
  `../bitrise-cli/internal/auth` (confirmed up to date at `310ac54` before porting;
  the 4 newer commits there only touch `rde`/`picker`, not `internal/auth`). `auth.go`
  copied verbatim (`Auth` struct, `IsOAuthManaged`, `TokenType`, `Path`/`Load`/`Save`/
  `Clear`, XDG-based `auth.yaml` path, atomic write via `.tmp`+rename, 0600/0700 perms);
  `auth_test.go`'s 9 cases ported with assertions adapted to this repo's testify
  (`require`/`assert`) convention instead of bitrise-cli's stdlib-only style, per
  `cli/cmdutil/modes_test.go`'s pattern. No new dependency: `gopkg.in/yaml.v3 v3.0.1`
  was already a direct, vendored dep here (used by `envfile/`) at the exact version
  bitrise-cli uses. Deliberately **not** wired into `cli/cli.go`'s `before()` yet —
  that's part of the next step (config handling) since `auth.Load()` only gets called
  once `config.Resolve` exists to take its result. Verification all green: `go build
  ./...`, `go vet ./...`, `go test ./internal/... -v` (9/9 pass), `go test ./...` (only
  the same pre-existing unrelated `TestParseAndValidatePluginFromYML` failure as
  BACKEND-573 above), `golangci-lint run ./...` (0 issues).
- **Next — extend config handling** (§7 step 2) — not started, no ticket yet (likely
  BACKEND-579 or split further, per the stack in §7). This is where `auth.Load()`
  actually gets wired into `cli/cli.go`'s `before()` and `config.Resolve` gets its
  `auth.Auth` param.
- Remaining §7 steps (API support, command groups one by one incl. `auth` login/logout,
  global flags both ways, analytics for new commands) — **later**.
