---
title: Bitrise CommandLine Interface Sharing
---

# Share your step with Bitrise CommandLine Interface

Now that you got CLI up and running and already created a step of your own there's not much left just publishing this awesome step and to let the world see how great job you've done!

`bitrise step share` can be used to share your steps. To be honest it's as simple as it gets. First you can check the help by simply running `bitrise step share` (no arguments) in the terminal. You'll see a colorful text and a step-by-step guide on how to share your Step.

## The process

Just a few words about sharing:
- Fork the StepLib repo you'd like to publish in
- Call `bitrise step share start --collection <your-fork-git-url>` to prepare your fork for publishing
- Add your Step with `bitrise step share create --git <step-git-url> --tag <version-tag> [--stepid <step-id>]`
- Call `bitrise step share finish` to upload the new Step yml to your StepLib fork
- Create a Pull Request and enjoy hundreds of thankful emails from developers around the world for making their life easier!
