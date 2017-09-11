# stepman

You can manage decentralized StepLib Step (script) Collections with `stepman`.

You can register multiple collections, audit them, manage the caching of individual Steps locally,
define and handle fallback URLs for downloading the Step codes (archives),
and share new Steps into public StepLib collections with `stepman`.

**Public Beta:** this repository is still under active development,
frequent changes are expected, but we we don't plan to introduce breaking changes,
unless really necessary. **Feedback is greatly appreciated!**

*Part of the Bitrise Continuous Integration, Delivery and Automations Stack,
with [bitrise](https://github.com/bitrise-io/bitrise) and [envman](https://github.com/bitrise-io/envman).*


## Install

Check the latest release for instructions at: [https://github.com/bitrise-io/stepman/releases](https://github.com/bitrise-io/stepman/releases)


## Share your own Step

Call `stepman share` and follow the guide it prints.

### Release a new version

1. Update go dependencies (`bitrise run godeps-update`)
1. PR & merge these changes to the `master` branch
1. Bump `RELEASE_VERSION` in bitrise.yml
1. Update the version test at: `./_tests/integration/version_test.go`
1. Commit (do not Push) these changes on `master` branch
1. Run `bitrise run create-release`
1. Fill the current version's `Release Notes` section in `CHANGELOG.md`
1. Push the changes to the `master` branch
1. Open the project's bitrise app on bitrise.io, find the triggered `create-release` workflow run's build
1. Download and test the generated bitrise binaries (`stepman version --full`)
1. Create the new version's release on [github](https://github.com/bitrise-io/stepman/releases/new):
  - Fill Tag and Version inputs
  - Copy paste the Changelog's `Release Notes` and `Install or upgrade` sections to the release description on github
  - Attach the generated (on bitrise.io) linux and darwin binaries to the release
  - Push the `Publish release` button on github