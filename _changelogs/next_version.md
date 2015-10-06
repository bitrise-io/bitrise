## Changes

* Bitrise setup now installs envman and stepman with `curl -fL` as it's the new recommended way. `curl -fL` will fail, if download was unsuccessfully.
* Fixed workflow triggering in pull request mode.
* Introduced new command: `bitrise share`, to share your step through bitrise.
