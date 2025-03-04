# Bitrise CLI

Bitrise CLI is the workflow runner that powers [Bitrise](https://bitrise.io/) builds. It's the component that runs inside build machines and execute steps defined in `bitrise.yml`.

It's also useful as a standalone dev tool in your local environment. You can:

- quickly validate your `bitrise.yml` changes before pushing a commit (`bitrise validate`)
- run CI workflows locally (`bitrise run workflow_name`)
- run the workflow editor in `localhost` and edit your configs and pipelines visually (`bitrise :workflow-editor`)
- perform various other tasks (for a full list run `bitrise help`)

## Install

<a href="https://repology.org/project/bitrise/versions">
    <img src="https://repology.org/badge/vertical-allrepos/bitrise.svg" alt="Packaging status" align="right">
</a>

There are multiple options to install Bitrise CLI:

- Homebrew: `brew install bitrise`
- Nix: packaged as `bitrise`, run `nix-shell -p bitrise` or your preferred configuration method.
- Download a pre-built binary from the [releases](https://github.com/bitrise-io/bitrise/releases) page
- There might be other [community-maintained packages](https://repology.org/project/bitrise/versions)

You can enable shell completion for the `bitrise run` command: [https://blog.bitrise.io/workflow-id-completion](https://blog.bitrise.io/workflow-id-completion)

### Building from source

First, set up the right Go version indicated by the `go.mod` file.

```sh
go install .
```

## Documentation

CLI documentation is part of the main [Bitrise docs](https://devcenter.bitrise.io). Relevant sections:

- [Workflows and Pipelines](https://devcenter.bitrise.io/en/steps-and-workflows.html)
- [Bitrise CLI local use](https://devcenter.bitrise.io/en/bitrise-cli.html)

## Tutorials and Examples

You can find examples in the [\_examples](https://github.com/bitrise-io/bitrise/tree/master/_examples) folder.

If you're getting started you should start with [\_examples/tutorials](https://github.com/bitrise-io/bitrise/tree/master/_examples/tutorials),
this should guide you through the basics, while you'll already use `bitrise` (requires installed `bitrise`).

You can find a complete iOS sample project at: https://github.com/bitrise-io/sample-apps-ios-with-bitrise-yml

## Tooling support & JSON output format

`bitrise` CLI commands support a `--format=[format]` parameter.
This is intended mainly for tooling support, by adding `--format=json` you'll
get a JSON formatted output on Standard Output.

Every error, warning etc. message will go to StdErr; and on the StdOut
you should only get the valid JSON output.

An example calling the `version` command:

`$ bitrise version --format=json`

Will print `{"version":"1.2.4"}` to the Standard Output (StdOut).

## Share your Step

You can use your own Step as you can see in the `_examples`, even if it's
not yet committed into a repository, or from a repository directly.

If you would like to share your awesome Step with others
you can do so by calling `stepman share` and then following the
guide it prints.
