# integrationtests/ - black-box suite

Separate Go module, deliberately outside `go test ./...` and deliberately not
vendored - do not run `go mod vendor` here. Tests spawn the real compiled binary as a
subprocess; nothing runs in-process.

Build tags: `//go:build linux_and_mac` for everything except the two Docker suites
under `docker/`, which are `linux_only`. One tag per file, with the legacy
`// +build` line.

## Running

```sh
go build -o /tmp/bitrise .          # from the repo root
bitrise setup                       # once per machine
cd integrationtests
INTEGRATION_TEST_BINARY_PATH=/tmp/bitrise go test --tags linux_and_mac -p 1 ./...
```

Docker suites run via `make docker-step-based-test` and `make docker-with-group-test`.

`-p 1` is required for a whole-suite run because `bitrise local run` shares machine
state across packages (steplib cache, `~/.bitrise`, temp and plugin dirs).

## Isolating a new test

- `cmd.AppendEnvs(...)`, not `t.Setenv`, so the test stays safe for `t.Parallel()`
- temp `XDG_CONFIG_HOME`, and clear `BITRISE_TOKEN` explicitly
- `cmd.SetDir(tmpDir)` so no ancestor `.bitrise-cli.yml` leaks in
- `BITRISE_ANALYTICS_DISABLED=true`

For a command that hits the API, serve canned responses from an `httptest.Server` and
point `api_base_url` (or `rde_api_base_url`) at it through a fixture `config.yml` in
that temp `XDG_CONFIG_HOME`. Loopback passes the https check.
