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

- merge every code changes to the master branch

- do not forget to merge every version related file changes:

  - update the version number (in version.go file)
  - update version tests (in _tests/integration/version_test.go file)

- push the new version tag to the master branch