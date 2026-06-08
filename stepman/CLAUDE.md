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

## Development Commands

This is a standard Go project, use standard Go tooling for development tasks.

`bitrise.yml` contains the tasks and workflows that run in CI

## Coding preferences

- Follow Go conventions as much as possible, and prefer simplicity and readability over cleverness.
- Robustness: This library is a critical part of the Bitrise stack, so reliability and robustnes are important. At the same time, avoid overly-defensive programming and unexpected fallback behavior, an explicit early failure is often better than a silent fallback.
- Most domain-specific structs are in `models/` (e.g. `Step`, `StepLib`, `VersionConstraint`). Before writing a new struct, check if it can fit into the existing models. It is not a hard requirement to put everything in `models/`, but it is a good starting point.
- Testing: Use `stretchr/testify` for assertions and test organization.
