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

## 7. v3 merge recipe (later task)

Prerequisite: port `bitriseapi/` (top-level), `internal/` (top-level), and the new
`cmd/cmdutil` helpers (into `cli/cmdutil`) first.

Per group `<g>` (`app`, `build`, `rde`+nested, `config`, `user`, `api`; plus new files for `yml`/`step`):
1. Copy the dir: `cp -R ../bitrise-cli/cmd/<g> cli/<g>` (for yml/step, copy only the new files).
2. Rewrite import prefixes in the copied files:
   - `…/bitrise-cli/cmd/<g>` → `…/bitrise/v2/cli/<g>`
   - `…/bitrise-cli/cmd/cmdutil` → `…/bitrise/v2/cli/cmdutil`
   - `…/bitrise-cli/bitriseapi` → `…/bitrise/v2/bitriseapi`
   - `…/bitrise-cli/internal/…` → `…/bitrise/v2/internal/…`
3. Wire root: add `rootCmd.AddCommand(<g>.NewCmd())` (or uncomment in yml/step cmd.go).
4. Vendor new deps: `go mod tidy && go mod vendor` (bubbletea, lipgloss, termenv,
   golang.org/x/term). Reconcile Go version (old `go.mod` vs new `go 1.26.2`).
5. Reconcile global flags (`--output/-o/--quiet/--no-color/--theme` vs `--debug/--ci/--pr`;
   watch `-o` collisions); make ported commands read `--output` via `cmdutil.ResolveFormat`.
6. Reconcile config/auth/DI: merged PersistentPreRunE keeps old `configs.InitPaths`/mode env
   AND adds new `config.Load`+`Resolve`+`WithResolved`+`style.Configure`.
7. Reconcile analytics: decide whether ported online commands call `cmdutil.LogCommandParameters`.
8. `go build ./... && go test ./cli/... && golangci-lint run`.

Known conflicts to decide (user-visible): config location `~/.bitrise` vs `~/.config/bitrise`;
auth token store (new); output/log (two systems — keep stdout=data / stderr=diagnostics so
JSON stays pipeable); self-update (two mechanisms — pick one); bare-parent default per group.

## 8. Task status

- BACKEND-571 — migrate old commands to cobra — **done**.
- BACKEND-573 — reorganize into per-group sub-packages (this task) — **in progress**.
  Executing against a generated exec spec (`reorg-spec.md`), step order: cmdutil
  extract → yml → step → plugin → local → rename root.go. Each step is gated on green
  `go build ./...` + `go test ./cli/...` before moving to the next (all green so far;
  no pre-existing test failures observed; no import cycles hit).
  - **Step 1 (cmdutil) — DONE.** Created `cli/cmdutil` (`flags.go`, `modes.go`,
    `command_analytics.go`, `args.go`, `json_output.go`, `failf.go`, `subcommand.go`
    (RequireKnownSubcommand/AsHidden/ShowSubcommandHelp), `run_modes.go` (Is*/Register*
    mode funcs), `config.go` (config/inventory helpers + CreateDefaultMerger +
    DefaultBitriseConfigFileName/DefaultSecretsFileName/OutputFormatKey), `update.go`
    (CheckUpdate/LatestTag/InstalledWithBrew/NewVersionFromBrew/NewCLIVersion/
    PrintCLIUpdateInfos). `modes_test.go` moved here (only tested these funcs).
    `globalTracker` is now an unexported cmdutil package var + exported
    `SetTracker`/`Tracker()`. All root `cli` files updated to call `cmdutil.*`;
    `cli.Run()` now does `tracker := analytics.NewDefaultTracker();
    cmdutil.SetTracker(tracker); defer tracker.Wait()`. `run_util.go` trimmed to just
    the WorkflowRunner engine; `depManagerBrew` const stays in root `run.go` for now
    (moves to `local` in step 5). `configShortKey`/`inventoryShortKey` kept unexported
    in cmdutil per spec; root files use literal `"c"`/`"i"` shorthand flags instead.
  - **Step 2 (yml) — DONE.** `cli/yml/{validate.go,merge.go,cmd.go}`, package `yml`,
    exported `NewValidateCommand`/`NewMergeCommand`/`NewCmd`. Old `cli/yml.go` deleted;
    root `commands.go` wires `yml.NewCmd()` + hidden aliases via
    `cmdutil.AsHidden(yml.NewValidateCommand())` etc.
  - **Step 3 (step) — DONE.** `cli/step/{cache.go,share.go,share_start.go,
    share_create.go,share_audit.go,share_finish.go,cmd.go}`, package `step`, exported
    `NewListCachedStepsCommand`/`NewPreloadStepsCommand`/`NewShareCommand`/`NewCmd`/
    `NewLegacyStepsCommand`. Old `cli/step.go` deleted; root wires `step.NewCmd()`,
    `cmdutil.AsHidden(step.NewShareCommand())`, `step.NewLegacyStepsCommand()`.
  - **Step 4 (plugin) — DONE.** `cli/plugin/{install.go,update.go,delete.go,info.go,
    list.go,cmd.go}`, package `plugin`. Converted the old `var pluginXCommand = &cobra.Command{...}` +
    `init()` pattern into `newInstallCommand()`/`newUpdateCommand()`/`newDeleteCommand()`/
    `newInfoCommand()`/`newListCommand()` funcs (unexported — only `NewCmd()` is called
    from root) called from `cmd.go`'s `NewCmd()`. `detectPlugin`/`runPlugin` moved out of
    old `plugin.go` into new root file `cli/plugin_dispatch.go` (still package `cli`,
    unchanged bodies) since they iterate `root.Commands()` and are pinned there by
    `cli_test.go` (`Test_detectPlugin`, `Test_envmanPassthrough` — both still pass). Old
    `cli/plugin.go` deleted; root `commands.go` imports `cli/plugin` as `plugin` and wires
    `plugin.NewCmd()` (no identifier collision — the local var named `plugin` in
    `runPlugin`/`detectPlugin` lives in `plugin_dispatch.go`, a different file that never
    imports the `plugin` package).
  - **Step 5 (local) — DONE.** Moved into `cli/local/` (package `local`): `cmd.go`
    (was `local.go`, now `NewCmd()` — mounts Run/Init/Setup/Tools/WorkflowList +
    `cmdutil.AsHidden(NewTriggerCommand())` for hidden `local trigger`), `run.go`,
    `init.go`, `setup.go`, `tools.go`, `workflow_list.go`, `trigger.go` (var+init
    converted to `NewTriggerCommand()` constructor — a `*cobra.Command` can only bind to
    one parent, and this one needs to mount under both `local` and hidden top-level),
    `run_config.go`, `run_trigger_params.go`, `build_run_result_collector.go`,
    `agent.go`, `step_activator.go`, `step_environment.go`, `analytics.go`,
    `run_util.go` (the WorkflowRunner engine), plus all their tests (`run_test.go`,
    `run_util_test.go`, `run_util_pipeline_test.go`, `run_util_validataion_test.go`,
    `run_trigger_params_test.go`, `step_environment_test.go`, `agent_test.go`,
    `analytics_test.go`, `tools_test.go`, `trigger_test.go`). Exported
    `NewRunCommand`/`NewInitCommand`/`NewSetupCommand`/`NewToolsCommand`/
    `NewWorkflowListCommand`/`NewTriggerCommand`/`NewCmd`. `depManagerBrew` const
    stayed with the engine (moved alongside `run.go`/`run_util.go`). Root `commands.go`
    now imports `cli/local` and wires `local.NewCmd()` + hidden top-level aliases
    (`cmdutil.AsHidden(local.NewRunCommand())` etc., incl.
    `cmdutil.AsHidden(local.NewTriggerCommand())` for the deprecated top-level `trigger`).
    `go test ./cli/local/...` green (138s, matches historical runtime — no behavior
    change, just package move).
  - **Step 6 (finalize) — DONE.** `git mv cli/commands.go cli/root.go`. Final gate all
    green: `go build ./...`, `go vet ./...`, `go test ./...` (one FAILING test,
    `TestParseAndValidatePluginFromYML` in `plugins/model_methods_test.go` — confirmed
    PRE-EXISTING via `git stash`: fails identically on the untouched tree because
    `stepman` binary isn't installed in this sandbox; nothing to do with the refactor).
    Manual spot-checks all passed: `go run . --help` shows local/yml/step/plugin/
    version/update/envman; `go run . run --help` and `go run . local run --help` both
    work; `go run . yml validate --help` works; `go run . step --help` and
    `go run . steps preload --help` both work; `go run . local trigger --help` and
    `go run . trigger --help` both work and `trigger` does NOT show up in
    `go run . local --help`'s command list (hidden as intended).
  - **BACKEND-573 reorg — DONE.** Final `cli/` tree: `cli/{cli.go,root.go,version.go,
    update.go,envman.go,help.go,cli_test.go,plugin_dispatch.go}` (root, package `cli`)
    + `cli/{cmdutil,local,yml,step,plugin,docker,containermanager}/` sub-packages. No
    import cycles. Not committed — left for the user to review/commit.
- v3 merge — port bitriseapi/internal/cmdutil, then per-group merge (recipe §7) — **later**.
