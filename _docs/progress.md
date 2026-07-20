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
- Jira: parent Story **BACKEND-568** "Merge into old CLI" tracks the effort overall;
  subtasks BACKEND-569..578 are an older, narrower breakdown, superseded in practice.
  **BACKEND-579** "Prepare old CLI for adding new commands" is the ticket actually
  tagged on every PR in the current stack — its own checklist (remove compat handlers
  → rename/alias → restructure into packages → auth token store → extend config
  handling → Bitrise API core → command groups) is the current source of truth for
  remaining scope (see §7-§8). Branch names still carry the historical
  `BACKEND-573-*` prefix — that's just naming, unrelated to which ticket the PRs are
  filed under.
- Execution is a stack of small, single-purpose PRs on a long-lived `v3-development`
  branch (merged to `master` later, when it makes sense), reviewed in order. See §8
  for current stack status.

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
| Config | `~/.bitrise/config.json` (JSON) | `~/.config/bitrise/cli/config.yml` + `auth.yaml` + per-dir `.bitrise-cli.yml` (XDG, YAML), precedence chain |
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
2. **Root stays `cli/`.** Root wiring stays in package `cli`; entrypoint stays
   `cli.Run()` — `main.go` unchanged.
3. **Shared helpers → `cli/cmdutil`** (package `cmdutil`), mirroring `cmd/cmdutil`.
   Leaf package (imports neither root nor groups) → no import cycle.
4. **Runner engine → `cli/local`.** The WorkflowRunner engine (step activation/env,
   build-result collector, agent hooks) is used only by `run` and `trigger`, so it
   lives in `cli/local`. `trigger` also moved to `cli/local` and is **hidden
   everywhere**: a hidden `local trigger` subcommand AND the hidden top-level
   `trigger` alias, both from one `local.NewTriggerCommand()` constructor.
5. **`bitriseapi/` top-level, `internal/` top-level** — created in the *later* v3
   merge, not now. `cmdutil` stays under `cli/` (the future `cmd/cmdutil` merges
   into it).
6. **Scope of the reorg = structure + docs only.** No bitrise-cli command/service/API
   code ported as part of it.

## 5. Reorg — landed

The flat `cli` package was split into per-group sub-packages (`cli/local`, `cli/yml`,
`cli/step`, `cli/plugin`) plus a leaf `cli/cmdutil` for shared helpers, per §4's
decisions. Landed across commits `c3b2356e` (removed legacy stuff), `06e04bae`
(trigger-check removal + hidden aliases), `dc569f05` (package split) — see the code
itself for the exact file layout and exported symbols rather than duplicating them
here.

Import graph (acyclic):
```
main.go → cli(root)
cli(root) → cli/local, cli/yml, cli/step, cli/plugin, cli/cmdutil
cli/{local,yml,step,plugin} → cli/cmdutil
cli/cmdutil → cli/docker, cli/containermanager, external only   (leaf)
```
Pre-cobra dispatch (`detectPlugin`, `runPlugin`, `envmanPassthrough`, `runEnvman`)
stays in package `cli` (`cli_test.go` pins it there) and works regardless of which
package built each command.

Old top-level names are re-registered as hidden aliases via `cmdutil.AsHidden`, each
built from the same exported leaf constructor as the group's own subcommand — a
`*cobra.Command` can only bind to one parent, so each alias needs its own fresh
instance from the constructor rather than reusing the group's command object.

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

## 7. v3 merge recipe (remaining steps)

Rather than redoing config/auth/API/flags/analytics reconciliation inside every
per-group port, those cross-cutting concerns are pulled out as their own upfront
steps. Order below is dependency-driven, checked against actual imports in
`../bitrise-cli` (not just the conceptual grouping) — see reasoning below the list.
May still be split into further sub-PRs as work proceeds:

1. ~~Port `internal/auth`~~ — **done**, see §8.
2. ~~Extend config handling~~ — **done**, see §8.
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

Why this order (checked via `grep` imports in `../bitrise-cli`, not assumed):
`internal/config` imports `internal/auth` (not the reverse) — auth had to land
first. `bitriseapi/` (incl. `rde/`) has no internal imports in either direction, so
it's not a hard prerequisite for config/auth at all — it only matters once a command
group needs it, hence splitting it into a small shared core now and per-domain files
later.

Known conflicts to decide (user-visible): config location `~/.bitrise` vs `~/.config/bitrise`;
auth token store (new); output/log (two systems — keep stdout=data / stderr=diagnostics so
JSON stays pipeable); self-update (two mechanisms — pick one); bare-parent default per group.

## 8. Task status

**Don't assume this section is current without re-reading it and re-checking
`gh pr list`/Jira — it changes every session.**

Current PR stack (all open, unmerged, as of 2026-07-02), tagged `[BACKEND-579]`:
- #1278 "Remove compatibility handlers" — base `v3-development`.
- #1279 "Rename and alias commands as described in the RFC" — base
  `BACKEND-573-legacy-cleanup`.
- #1282 "Restructure commands code into packages per command family, like Platform
  CLI" — base `BACKEND-573-command-reorganization`.
- #1284 "Add auth token store from Platform CLI (dependency of config handling)" —
  base `BACKEND-573-command-folders`. This is §7 step 1.
- Not yet a PR: config handling (§7 step 2) is committed on
  `BACKEND-579-config-handling`, stacked on top of #1284's branch.

Separately, BACKEND-571 (migrate urfave→Cobra) is **done**, PR approved but not yet
merged (Mobile team owns the phased rollout).

One unrelated pre-existing test failure throughout this whole stack:
`TestParseAndValidatePluginFromYML` (`plugins/model_methods_test.go`) fails because
`stepman` isn't installed in this sandbox — confirmed via `git stash` that it fails
identically on the untouched tree.

**§7 step 1 (auth token store, PR #1284) — done**, commit `9b92ee58`. Added
`internal/auth/{auth.go,auth_test.go}` (this repo's first top-level `internal/`
package), ported verbatim from `../bitrise-cli/internal/auth`. No new dependency
(`gopkg.in/yaml.v3` was already vendored).

**§7 step 2 (config handling) — done**, on `BACKEND-579-config-handling`. Scope was
narrowed during planning: bitrise-cli's 7-key config schema (`app_id`, `token`, API
URLs, `output`, `theme`) has no consumer in this repo yet, so only
`configs.ConfigModel`'s three real fields (`SetupVersion`, `LastCLIUpdateCheck`,
`LastPluginUpdateChecks`) got the layered treatment.
- New `internal/config/{config.go,resolve.go}`: reads `~/.config/bitrise/cli/config.yml`
  (XDG) + per-dir `.bitrise-cli.yml` (ancestor search). `Resolve(legacy, dir, global)`
  precedence: legacy `~/.bitrise/config.json` > per-dir > global > zero value — the
  RFC calls for reading the legacy JSON first if it exists, so it's authoritative
  over the new layers rather than a fallback beneath them; threaded via
  `cmd.Context()` (`WithResolved`/`FromContext`), same pattern as bitrise-cli.
- `configs.LoadConfigModel()` — new exported wrapper around the existing private
  `loadBitriseConfig()`; purely additive, existing setup/update-check functions
  untouched.
- **Known gap:** the RFC says "the new location will be used when saving," but
  `SaveSetupSuccessForVersion`/`SaveCLIUpdateCheck`/`SavePluginUpdateCheck` still
  write only to the legacy `~/.bitrise/config.json` — nothing writes to the new
  `config.yml` yet. Since legacy stays top-precedence for reads, this doesn't cause
  visibly wrong behavior for existing users, but new/first-time users never get a
  `config.yml` written, and the RFC's save-target behavior isn't implemented.
  Deliberately deferred to a later step (not fixed here).
- Wired into `cli/cli.go`'s `before()`. Load failures are non-fatal (logged +
  zero-value fallback) since nothing consumes the new config yet and all callers of
  `before()` otherwise treat a returned error as fatal.
- Consequence: `internal/auth` stays unwired (no field needs a token yet) — wire it
  once a real key needing one lands (§7 step 4).
- Verified: `go build/vet/test ./...` green (only the pre-existing stepman failure
  above); `golangci-lint run` 0 issues; manual check that a pre-existing legacy value
  wins over conflicting per-dir/global values, and that per-dir still overrides global
  when no legacy value is set.

**Resolved:** "layer the existing keys" meant `configs.ConfigModel`'s three fields
(tied to the literal old config file named in the original request), confirmed
against the RFC — not the alternative reading of extending `configs`' env-var-only
boolean toggles (`CI`/`DEBUG`/`PR`/secret filtering/offline mode) to also be
YAML-configurable, which stays out of scope for this step.

Remaining: §7 steps 3-6 (bitriseapi core, command groups incl. `auth` login/logout,
global flags both ways, analytics for new commands).
