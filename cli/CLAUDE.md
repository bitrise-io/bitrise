# cli/ - the command layer

Flags, arg validation, identifier resolution via `cmdutil`, rendering. No business
logic, no HTTP.

`cli/stack/list.go` is the example to copy. **`cli/rde` has its own conventions and
its own `cli/rde/CLAUDE.md` - read it before touching that tree.** Adding a command
usually means adding a service method too; `internal/CLAUDE.md` covers that half.

## Anatomy

Group `cmd.go` exports `NewCmd()`, leaves export `NewXxxCommand()` (rde uses
unexported `newXxxCmd()` instead), `cli/root.go` wires groups onto the root. A group
that only dispatches uses `RunE: cmdutil.RequireKnownSubcommand`.

Leaves use `RunE`, never `Run`, and always set `Args`:

- `cobra.NoArgs` when it takes none
- `cmdutil.RequireArgs("STACK_ID")` when the positional arg is the only identifier
- `cobra.MaximumNArgs(1)` when a flag, env var or config can supply it instead
  (`app view`)

`RequireArgs` enforces a lower bound only. It rejects no extra argument (nothing caps
one today) and checks no values, so guard an empty string in `RunE` where it is
meaningless, as `cli/rde/session/create.go` does. `build view` uses
`cobra.ExactArgs(1)` where `RequireArgs` would fit; follow the rule, not that file.

RunE body, in this order:

```go
cmdutil.LogCommandParameters(cmd)

format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
if err := output.ConfigureOutputFormat(format); err != nil {
    return fmt.Errorf("failed to configure output format: %w", err)
}

client, err := cmdutil.NewAPIClient(cmd)
if err != nil {
    return err
}

// resolve identifiers via cmdutil, then call the service
result, err := internalstack.NewService(client).List(cmd.Context(), workspace)
if err != nil {
    return fmt.Errorf("listing stacks failed: %w", err)
}

return output.Render(cmd.OutOrStdout(), output.Format, result,
    func(w io.Writer, result internalstack.StacksResult) error {
        return printStacksTable(w, result.Items)
    })
```

## Errors and streams

- Return errors, wrapped with `%w`. The root sets `SilenceErrors`/`SilenceUsage` and
  `cli/cli.go` hands the result to `cmdutil.Failf`.
- `cmdutil.Failf` and `log.Fatalf` exit from inside `RunE` and make it untestable.
  Legacy only, no new callers.
- `cmdutil.SilenceRootErrors` only when the command already printed its own summary.
- stdout carries the answer, stderr the diagnostics. Use `cmd.OutOrStdout()` and
  `cmd.ErrOrStderr()`, never the global logger, whose writer is stdout and would
  corrupt `... --format json | jq`.
- Gate human-only notices on the format, as `cli/stack/list.go` does with
  `output.Format == output.FormatRaw`.
- Capture every `fmt.Fprint*` return to a real stream, or use `cmdutil.NewErrWriter`
  and check `ew.Err` once. Writes into a `strings.Builder` are exempt.

## Rendering

Lists use `internal/style.Table`. Detail commands compose into a `strings.Builder` and
flush once:

```go
func printStackText(w io.Writer, st internalstack.Stack) error {
	var b strings.Builder
	fmt.Fprintf(&b, "ID:     %s\n", st.ID)
	_, err := io.WriteString(w, b.String())
	return err
}
```

Name it `print<Thing>Text`, keep it below the constructor in the same file, and move it
to the group's `utils.go` only when a second command renders the same type
(`cli/build/utils.go`). A resource with a page on app.bitrise.io also gets `--web`
(`app view`, `build view`); stacks have none.

## Flags

- `--output`/`-o` is root-persistent; a per-command `--format` wins. Values: `raw`
  (default, `human` alias), `json`, `yml`.
- `-f` is `--format` except where the command has `--file` (`yml update`,
  `yml validate`, `rde template create/update`), `--field` (`api`, no `--format` at
  all) or `--follow` (`rde session logs`).
- Use the constants: `cmdutil.FlagWorkspace`, `FlagApp`, `FlagOutput`, `FlagQuiet`,
  `FlagNoColor`, `FlagTheme`, `FormatKey`. Each doc comment covers precedence and env
  var. `cmdutil.IsQuiet(cmd)` reads `--quiet`.

## Vocabulary and help

- **ID, never slug** in flags, metavars, table headers, labels, config keys and JSON
  fields. `slug` stays in the API client and internal identifiers. The only user-facing
  `SLUG` is `$BITRISE_APP_SLUG`, which Bitrise sets in every build.
- **Workspace**, never organization, org or owner. `rde template`'s `OWNER` column is
  the creator's email, a different thing.
- Singular nouns. CRUD verbs: `create`, `update`, `delete`, `list`, `view`. `build`
  (and `local`) uses `trigger` instead of `create` to start a run, and `abort`
  instead of `cancel` to stop one.
- `Use`, `Short` and `Example` always. Add `Long` only for an example, a precedence
  rule or a surprise - cobra already prints the flags.

## Tests

Same package as the file under test. No mocking framework - fake the API with
`httptest.NewServer`.

Build the real command via its constructor, swap streams with `SetOut`/`SetErr`,
inject config with `cmd.SetContext(config.WithResolved(...))`.
`cli/stack/list_test.go` is the bootstrap to copy, including the two traps: point
`XDG_CONFIG_HOME` at a temp dir, and blank `BITRISE_TOKEN`, which outranks any
fixture. Cover the error path, not just the happy one.

## Hidden legacy aliases

`cli/root.go` registers `run`, `init`, `setup`, `tools`, `workflows`, `trigger`,
`validate`, `merge`, `share` and `steps` as hidden aliases so v2 command lines keep
working. Leave them; add no new ones.
