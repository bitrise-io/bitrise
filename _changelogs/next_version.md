## Changes

* BITRISE build result log improvements:
    * step run summary contains step version, and update note, if new version available
    * build run summary step sections contains step version, and update note, if new version available
* __BREAKING/FIX__ : `bitrise trigger` will **NOT** select any workflow in Pull Request mode if the pattern does not match any of the `trigger_map` definition.
* unified `config` and `inventory` flag handling: you can specify paths with `--config` and `--inventory`, and base64 encoded direct input with `--config-base64` and `--inventory-base64`. Can be used by tools, to skip the need to write into files.
* __FIX/BREAKING__ : environment handling order : App Envs can now overwrite the values defined in inventory/secrets (in the last version the secrets/inventory could overwrite the App Envs).
* `validate` command accepts `--format` flag: `--format=[json/raw]` (default is `raw`)
* new command: `step-list` (lis of available steps in Step Lib) `bitrise step-list`
* new command: `step-info` (infos about defined step) `bitrise step-info --id script --version 0.9.0`
* revision of `normalize`, to generate a better list/shorter output list
