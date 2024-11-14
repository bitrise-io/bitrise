# Bitrise CLI

_Discussion forum: [https://discuss.bitrise.io/](https://discuss.bitrise.io/)_

Run your Bitrise automations with this CLI tool on any Mac or Linux machine, and use the same configuration on
[bitrise.io](https://www.bitrise.io) (automation service, with a mobile app focus).

_Part of the Bitrise Continuous Integration, Delivery and Automations Stack,
with [stepman](https://github.com/bitrise-io/stepman) and [envman](https://github.com/bitrise-io/envman)._

For a nice & quick intro you should check: [https://www.bitrise.io/cli](https://www.bitrise.io/cli)

## Install and Setup

The installation is quick and easy, check the latest release for instructions at: [https://github.com/bitrise-io/bitrise/releases](https://github.com/bitrise-io/bitrise/releases)

Installing with Homebrew:

`brew update && brew install bitrise`

Optionally, you can call `bitrise setup` to verify that everything what's required for `bitrise` to run
is installed and available, but if you forget to do this it'll be performed the first
time you call `bitrise run`.

You can enable shell completion for the `bitrise run` command: [https://blog.bitrise.io/workflow-id-completion](https://blog.bitrise.io/workflow-id-completion)

## Tutorials and Examples

You can find examples in the [\_examples](https://github.com/bitrise-io/bitrise/tree/master/_examples) folder.

If you're getting started you should start with [\_examples/tutorials](https://github.com/bitrise-io/bitrise/tree/master/_examples/tutorials),
this should guide you through the basics, while you'll already use `bitrise` (requires installed `bitrise`).

You can find a complete iOS sample project at: https://github.com/bitrise-io/sample-apps-ios-with-bitrise-yml

## Tooling support & JSON output format

`bitrise` CLI commands support a `--format=[format]` parameter.
This is intended mainly for tooling support, by adding `--format=json` you'll
get a JSON formatted output on Standard Output.

**This is still work-in-progress, we're working on providing
the `--format` param to every command except `run`**.

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

## Documentation

We added some documents to make it a bit easier to get started with Bitrise CLI. The documentation includes a quick and a little longer guides for CLI, a [React Native](http://facebook.github.io/react-native/) project workflow guide and an overview of the Step share process. You can find them in the [\_docs](/_docs/) folder.

## Development

### Guidelines

* **Easy to use**: the UX for the end-user, always keep it in mind, make it a pleasant experience to work with this tool (and all of the Bitrise tools)!
* **Code should be kept simple**: easy to understand, easy to collaborate/contribute (as much as possible of course).
* **Compatibility**: never do an incompatible change, unless you can't avoid it. Release new features as additional options, to not to break existing configurations.
* **Stability**: related to compatibility, but in general stability is really important, especially so in a CI/automation environment, where you expect fully reproducible outcomes.
* **Flexibility**: should also be kept in mind, but only if it does not affect the previous points.

### Updating dependencies

To do a full dependency update use [bitrise-tools/gows](https://github.com/bitrise-tools/gows),
for a clean workspace:

```
gows clear && gows bitrise run dep-update
```

to test that all dependency is included:

```
gows clear && gows go test ./... && gows go install && gows bitrise run test
```

and/or with `docker-compose`:

```
docker-compose build && docker-compose run --rm app go test ./...
```

### Local Dev Workflow
The following commands will work to get you started using a text editor such as VCSCode or similar. 

All commands should be run from the root directory.

* Setup:

    ```mkdir -p ./_bin```

* Build:

    ```go build -o ./_bin```

* Run:
    
    ```./_bin/bitrise COMMAND```

