---
title: Running React Native projects with Bitrise CommandLine Interface
---

# Running React Native projects with Bitrise CommandLine Interface

Check out our [sample workflow](../_examples/tutorials/react-native/bitrise.yml) that uses React Native. Some of the used variables were added to the `.bitrise.secrets.yml` before starting to run the workflow add the variables specific to your application. The list of variables:

- REPO_URL:
- webhook_url:
- channel:
- from_username:
- message:
- from_username_on_error:
- message_on_error:

Now that you configured your .yml let's see what is in the bitrise.yml file.

We presume you are familiar with the bitrise.yml structure from the introduction so jump to the `run-react-native` workflow.

There are four steps:
- The first script is a simple git clone or pull if the source code has already been downloaded. There's no need to make your funky music a bit laggy or just simply take the precious time from seeing a successful build log with deleting and cloning it again.
- Next we have another script that configures React Native. You can also check our [guide on our DevCenter](http://devcenter.bitrise.io/tutorials/building-react-native-projects-on-bitrise.html), this script was created from that guide. After this step you can run every Xcode related step.
- Now we have the project and installed React Native, all we have to do is run the [new-xcode-test](https://github.com/bitrise-io/bitrise-steplib/tree/master/steps/new-xcode-test/0.9.1).
- Well to be honest that's all. The final step is a Slack message. It sends a message to a given channel. The message depends on the success of the build.

So as a quick overview let's go through it again. We have a clone, a React Native setup, an Xcode Test and a notification step. And we're done, it's that simple! Feel free to try it with your own project and if you get stuck contact us!
