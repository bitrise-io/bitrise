# cli/rde/ - Remote Dev Environment commands

This tree diverges from `cli/CLAUDE.md` and the other cloud groups for historical reasons.

- **Inside `cli/rde/`**: match what is here. These commands have users and the
  divergences are part of the behavior they depend on. Do not correct them toward
  `cli/CLAUDE.md`.
- **Outside `cli/rde/`**: follow `cli/CLAUDE.md`. Nothing here is a pattern to spread.

Neither rule licenses a new exception. Copy `session/list.go` for a read or
`session/update.go` for a mutation - **not `cli/stack/list.go`, whose differences are
invisible in a diff, so the result builds fine and is wrong for this directory.**

The list below is what differs deliberately. Anything not on it follows
`cli/CLAUDE.md`, including where existing files here do not: several commands have no
`Example`, for instance, and new ones still need one.

## What differs

- **Leaf constructors are unexported**: `newViewCmd()`, not `NewViewCommand()`. Only
  the group's `NewCmd()` is exported.
- **`cmdutil.RequireArgs("SESSION_ID")` instead of `cobra.ExactArgs`**, which names the
  missing argument. It sets a lower bound only and checks no values, so guard an empty
  string in `RunE`, as `session/create.go` does.
- **A bare group root runs `list`.** The `session`, `stack`, `machine-type`,
  `saved-input` and `template` roots set `Args: cobra.NoArgs` +
  `RunE: cmdutil.DelegateToList`. Leaves are unaffected. Every other group root,
  including `bitrise rde` itself, prints help via `RequireKnownSubcommand`.
- **`--format` is bound.** `var format string` +
  `StringVar(&format, cmdutil.FormatKey, ...)`, then use the variable. The rest of the
  CLI leaves it unbound and calls `GetString` in `RunE`. No file crosses over.
- **`--workspace` is an ID.** `cmdutil.ResolveWorkspaceID` returns an explicit flag
  value verbatim. `app` and `stack` use `ResolveAndLookupWorkspaceSlug`, which accepts
  a display name. With flag, env and default all empty, `ResolveWorkspaceID`
  auto-detects a sole workspace or prompts via `cli/cmdutil/picker`.
- **An impossible format is rejected, not rerouted:**
  `fmt.Errorf("--watch cannot be combined with --format %s (it re-renders continuously, not a single-object result)", output.Format)`.
  So do `session logs` and `session vnc --forward`. `session exec` keeps `--format`
  because it has a terminal object. Outside this tree, `build watch` keeps the flag
  and moves the stream to stderr instead.
- **Banners gate on quiet too:** `if !cmdutil.IsQuiet(cmd) && output.Format == output.FormatRaw`.
  `--quiet` is root-persistent but read only here and in `ResolveWorkspaceID`.
- **`-f` is `--follow` on `session logs`**, which registers `--format` with no
  shorthand.
- **Tests isolate once per package** via a `main_test.go` holding only
  `func TestMain(m *testing.M) { os.Exit(cmdtest.RunIsolated(m)) }`, then run through
  `cmdtest.Run`. Other groups use per-test `t.Setenv`. Do not import `cli/cmdtest`
  elsewhere.
- **Local state** beyond `config.yml`/`auth.yaml`: `internal/rde/localsession` keeps
  resumable sessions and preferences under `<config dir>/rde/projects/<repo path>/`.

## What does not differ

A change to any of these is a plain bug, not a local convention:

- stdout carries the answer, stderr the diagnostics and prompts
- `-f` belongs to `--file` on `template create`/`update`, as in `yml update`
- shared `cmdutil.ReadSecretInput` / `ReadPasswordInput` / `CheckValueStdinPiped`
- `--format` registered per command, never inherited
- `RunE` always, `Run` never, `Args` always set
- layering: `cli/rde` calls `internal/rde` calls `internal/rdeapi`. The separate client
  is deliberate - RDE has its own backend and base URL, via `cmdutil.NewRDEClient`.
