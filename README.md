# Bitrise CLI

[![Join the chat at https://gitter.im/bitrise-io/bitrise-cli](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/bitrise-io/bitrise-cli?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

**Technology Preview:** this repository is still under active development, breaking changes are expected and feedback is greatly appreciated!

Bitrise (offline) CLI

You can run your Bitrise workflow with this CLI tool,
on your own device.

Part of the Bitrise Continuous Integration, Delivery and Automations Stack.


## Install and Setup

To install `bitrise-cli`, run the following commands (in a bash shell):

```
curl -L https://github.com/bitrise-io/bitrise-cli/releases/download/0.9.1/bitrise-cli-`uname -s`-`uname -m` > /usr/local/bin/bitrise-cli
```

Then:

```
chmod +x /usr/local/bin/bitrise-cli
```

The first time it's mandatory to do a `setup` as well after the install,
and as a best practice you should the a setup every time you install a new version of `bitrise-cli`.

Doing the setup is as easy as:

`bitrise-cli setup`

Once the setup finishes you have everything in place to use `bitrise-cli`.


## Development Guideline

* Number one priority is UX for the end-user, to make it a pleasant experience to work with this tool!
* Code should be kept simple: easy to understand, easy to collaborate/contribute (as much as possible of course).


## Tests

* Work with multiple projects, in separate folders
