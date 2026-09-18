# internal/ - services, API clients, stored state

What `cli/` may not do: HTTP, business rules, and the files the CLI persists.
`cli/CLAUDE.md` covers the command half of the same change.

## Service layer

`NewService(client)`, `ctx` first, CLI-shaped result structs rather than wire types.
Methods are named for the command verb (`List`, `View`, `Get`, `Update`, `Trigger`,
`Abort`, `Validate`), but the method set follows the API: one per operation, not one
per command. A command whose work an existing method already does calls it instead of
gaining a wrapper that only forwards.

- A collection returns a wrapper for a stable top-level object,
  `StacksResult{Items []Stack}`; a single-item fetch returns the bare item.
  `internal/rde` predates the rule and returns bare slices, wrapping in the command.
- Normalize the wire shape here, not in the command: `stack-report` becomes
  `stack_report`. `TestStacksResult_JSONShape` pins it.
- Nothing found returns an error.

**IMPORTANT: every struct that can reach `output.Print` needs a `yaml` tag beside each
`json` tag** - same name, same `omitempty`:

```go
ID        string `json:"id" yaml:"id"`
OSVersion int    `json:"os_version,omitempty" yaml:"os_version,omitempty"`
```

`gopkg.in/yaml.v2` ignores `json` tags and falls back to the lowercased field name, so
a missing tag makes `--format yml` emit `osversion` while `--format json` emits
`os_version`. The code compiles, the tests pass, and the output is wrong.

## Clients

- `internal/bitriseapi` - bearer token, generic `get`/`getEnvelope`/`getPage`/
  `postDecode`, one file per resource, failures as `*bitriseapi.APIError`.
  `New` returns an error because it validates its base URL.
- `internal/rdeapi` - same role for the RDE backend.
- `internal/webclient` - short-lived cookie/CSRF client for the Rails sign-up and
  email-password endpoints. Never persisted.

The client covers only endpoints a command has needed, so a per-item GET may be
missing where the resource obviously has one - stacks are fetched only as a full list.
The repo carries no API spec, so **never add a client method for an endpoint whose
existence has not been established; ask first.** Falling back to the bulk endpoint is
legitimate: filter in the service, not the command, and note in the doc comment that
the caller can no longer tell "absent" from "not visible to this token".

## Config, auth, locking

- `internal/config` resolves layers once and threads `Resolved` through `context`
  (`config.WithResolved` / `config.FromContext`). Precedence: legacy
  `~/.bitrise/config.json` > per-dir `.bitrise-cli.yml` > global
  `~/.config/bitrise/cli/config.yml` > default. Writes are atomic.
- `APIBaseURL`, `WebBaseURL` and `RDEAPIBaseURL` skip the per-directory layer, so a
  cloned repo cannot redirect where the token is sent. `AppID`, `DefaultWorkspaceID`,
  `Output` and `Theme` honor it. New credential-relevant keys follow the URLs.
- `internal/auth` keeps the token in `auth.yaml` beside `config.yml` at 0600.
  `BITRISE_TOKEN` wins over the file. A token with a refresh token is OAuth-managed
  and refreshes through `cmdutil.OAuthConfig()`.
- `internal/baseurl.Validate(label, raw)` requires https except on loopback. Every URL
  that will carry a credential goes through it.
- `internal/filelock.Lock` is the shared cross-process lock; `internal/auth` and
  `internal/config` differ only in their durations. Any read-modify-write on a file
  crossing invocations uses it. Mechanism tests live in `internal/filelock`; a wrapper
  only proves it passes the right options.
- `internal/style` is for human output only - ANSI must never reach JSON or YAML.
  Pass the real writer to `style.New(w)`; a throwaway buffer looks non-interactive and
  silently disables color.
