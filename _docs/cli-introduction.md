---
title: Bitrise Command Line Interface introduction
---

# Installing Bitrise Command Line Interface

For a more detailed overview see the [CLI how to guide](cli-how-to-guide.md)
Let's cut to the chase! Our Command Line Interface is now available via [Homebrew](https://github.com/Homebrew/homebrew/tree/master/share/doc/homebrew#readme) so first call that good old `brew update` just to be sure and when it's done simply call the `brew install bitrise` command in your terminal. And BOOM! you can start using it right away!

If you choose to go the old fashioned way just check the [releases site](https://github.com/bitrise-io/bitrise/releases) for instructions.

## Setting up Bitrise

The installation is done so let's run the `bitrise setup` command to finish up and install the missing dependencies!

## Create your first project

Let's run the `bitrise init` command in the terminal. Your first Workflow is ready to be run!

Make sure you are in the current project's directory and run the `bitrise run` command. It will show you the workflows listed in the bitrise.yml. Now you simply have to choose one (after the init there's only one called `primary`) from the list and call `bitrise run <workflowname>` and watch CLI execute your Workflow Step-by-Step! You can add Steps to the Workflow from our [StepLib](https://github.com/bitrise-io/bitrise-steplib/tree/master/steps) or even from your own StepLib fork if you have one.

Happy Building!
