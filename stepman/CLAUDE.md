# CLAUDE.md

## Project Overview

Stepman is the tool behind Bitrise steps: it allows submitting new steps to the step library, as well as running steps in a workflow. It's a Go library primarily (imported by the Bitrise CLI), but it also has a CLI interface for step management and sharing.

## Architecture

- **Core Models**: Step definitions, library management, and version constraints in `models/`
- **Library Management**: StepLib operations and caching logic in `stepman/`
- **Step Activation**: Step execution and toolkit handling in `activator/`
- **Toolkits**: Support for different step types (Bash, Go, Swift) in `toolkits/`
- **Step ID Management**: Step identification and versioning in `stepid/`

Key entry point is `main.go` → `cli.Run()` which sets up the CLI application with commands defined in `cli/commands.go`.

### Step activation paths

Activation goes through `activator.Activator`, built once per run with, so every step of a run shares one HTTP client.

- **Legacy git-clone path**: clones the steplib repo locally and reads spec.json. This is the codepath used for custom/self-hosted steplibs.
- **StepLib API path** (default): fetches static JSON over HTTPS (no git clone).

Precompiled step executables (skip building from source) are controlled by `Options.UsePrecompiled`, with the storage URLs to try in `Options.PrecompiledStorageURLs`.

## Development Commands

This is a standard Go project, use standard Go tooling for development tasks.

`bitrise.yml` contains the tasks and workflows that run in CI

## Releasing

Pushing a `vX.Y.Z` git tag on `master` triggers the release: the Tooling Control Center Bitrise project runs the `binary-tool-release` workflow (Goreleaser), builds binaries, and creates a GitHub release. Consumers (e.g. the Bitrise CLI) then bump the module version. Note: `v0.22.0` is a poisoned tag serving a stale API — never depend on it; use `v0.23.0+`.

## Coding preferences

- Follow Go conventions as much as possible, and prefer simplicity and readability over cleverness.
- Robustness: This library is a critical part of the Bitrise stack, so reliability and robustnes are important. At the same time, avoid overly-defensive programming and unexpected fallback behavior, an explicit early failure is often better than a silent fallback.
- Most domain-specific structs are in `models/` (e.g. `Step`, `StepLib`, `VersionConstraint`). Before writing a new struct, check if it can fit into the existing models. It is not a hard requirement to put everything in `models/`, but it is a good starting point.
- Testing: Use `stretchr/testify` for assertions and test organization.
