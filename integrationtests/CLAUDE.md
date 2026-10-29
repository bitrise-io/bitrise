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
INTEGRATION_TEST_BINARY_PATH=/tmp/bitrise go test --tags linux_and_mac ./...
```

Docker suites run via `make docker-step-based-test` and `make docker-with-group-test`.

Packages run in parallel on one machine, so a test that changes state other tests read
breaks them. The rules below keep such a test's changes to itself.

## Isolating a new test

- `cmd.AppendEnvs(...)`, not `t.Setenv`, so the test stays safe for `t.Parallel()`
- temp `XDG_CONFIG_HOME`, and clear `BITRISE_TOKEN` explicitly
- `cmd.SetDir(tmpDir)` so no ancestor `.bitrise-cli.yml` leaks in
- `BITRISE_ANALYTICS_DISABLED=true`
- CI runs the tests without `ENVMAN_ENVSTORE_PATH`, so they can't write into the build's
  envstore; a test that needs an envstore creates its own and passes it with `cmd.AppendEnvs`
- a test that changes `~/.stepman`, `~/.bitrise` or `~/.config/bitrise` (steplib cache,
  agent config, toolkit caches, the record of the version set up) runs bitrise with
  `HOME=testhelpers.IsolatedHome(t)`
- a test that changes the binary itself runs a copy of it, as `Test_Update` does

For a command that hits the API, serve canned responses from an `httptest.Server` and
point `api_base_url` (or `rde_api_base_url`) at it through a fixture `config.yml` in
that temp `XDG_CONFIG_HOME`. Loopback passes the https check.
