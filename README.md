# Bitrise (offline) CLI

Run your Bitrise automation workflows with this CLI tool on your own development device, and on
your Continuous Integration tool / service.

*Part of the Bitrise Continuous Integration, Delivery and Automations Stack,
with [stepman](https://github.com/bitrise-io/stepman) and [envman](https://github.com/bitrise-io/envman).*


## Install and Setup

The installation is quick and easy, check the latest release for instructions at: [https://github.com/bitrise-io/bitrise/releases](https://github.com/bitrise-io/bitrise/releases)


## Tutorials and Examples

You can find examples in the [_examples](https://github.com/bitrise-io/bitrise/tree/master/_examples) folder.

If you're getting started you should start with [_examples/tutorials](https://github.com/bitrise-io/bitrise/tree/master/_examples/tutorials),
this should guide you through the basics, while you'll already use `bitrise` (requires installed `bitrise`).

You can find a complete iOS sample project at: https://github.com/bitrise-io/sample-apps-ios-with-bitrise-yml


## Development Guideline

* Number one priority is UX for the end-user, to make it a pleasant experience to work with this tool!
* Code should be kept simple: easy to understand, easy to collaborate/contribute (as much as possible of course).
* Flexibility should also be kept in mind, but only if it does not affect the previous two points.


## Share your Step

You can use your own Step as you can see in the `_examples`, even if it's
not yet committed into a repository, or from a repository directly.

If you would like to share your awesome Step with others
you can do so by calling `stepman share` and then following the
guide it prints.

## Documentation

We added some documents to make it a bit easier to get started with Bitrise CLI. The documentation includes a quick and a little longer guides for CLI, a [React Native](http://facebook.github.io/react-native/) project workflow guide and an overview of the Step share process. You can find them in the [_docs](/_docs/) folder.
