## Changes

* __BREAKING__ / __FIX__ : If `bitrise trigger` is called with a trigger pattern that doesn't match any expression in `trigger_map` the workflow with the same name as the trigger pattern name **will no longer be selected**. This feature proved to have more issues than pros.
* __DEPRECATION__ : the previous `dependencies` property is now deprecated. From now on, dependencies should be declared in the `deps` property which has a new syntax, grouped by dependency managers. The syntax will be extended in the near future but in a backward compatible way.
  Supported dependency managers: `brew`, `apt_get` and `check_only` (for checking dependencies which can't be installed automatically, ex the `Xcode.app`).
  Example:

  ```
  - step:
      deps:
        brew:
        - name: cmake
        - name: git
        - name: node
        apt_get:
        - name: cmake
        check_only:
        - name: xcode
  ```
* Improved validate command output.
* __BREAKING__ : if you don't specify the version of a step `bitrise` will now try to update the local Step Lib cache before using the Step. Previously the latest version **available in the local cache** was used, but this caused more confusion. The local cache is still used in this case if the Step Lib can't be updated, it'll still work in case of a network issue.
* __BREAKING__ : From `bitrise step-info` the `--id` flag was removed, the first cli param used as step id, no need to write the `--id` flag anymore. Example: `bitrise step-info script` instead of `bitrise step-info --id script`.
* __IMPORTANT__ : `format_version` bumped to `1.1.0`, which means that the `bitrise.yml` generate with this `bitrise` version won't be compatible with previous `bitrise` versions. Previous `bitrise.yml`s of course still work with this new `bitrise` version.
* Now you can use all your environment variables (secrets, app, workflow envs and step outputs) in `run_if` fields and in step inputs.
* `bitrise step-info` got a new option `--step-yml` flag, which allows printing step info from the specified `step.yml` directly (useful for local Step development).
* Step inputs got a new field: `IsTemplate` / `is_template`. This field indicates whether the value contains template expressions which should be evaluated before using the value, just like in case of `is_expand`. The template expression have to be written in Go's template language, and can use the same properties as `run_if` templates can. Example:

  ```
  - script:
    title: Template example
    inputs:
    - content: |-
        {{if .IsCI}}
        echo "CI mode"
        {{else}}
        echo "not CI mode"
        {{end}}
      opts:
        is_template: true
  ```
* Improved environment and input value and options casting:
    * Now you can use `"NO"`, `"No"`, `"YES"`, `"Yes"`, `true`, `false`, `"true"`, `"false"` in every place `bitrise` expects a bool value (ex: `is_expand`).
    * Every field where `bitrise` expects a string in now casted into a string. This means that you can now use `true` and `false` instead of `"true"` and `"false"` in `value_options`. Same is true for the input and environments value itself, so you can now write `true` instead of `"true"` and it'll still be casted to string.
* Pull Request and CI mode handling extension: related flag environment variables can now be defined in `secrets` / `inventory` as well.
* `bitrise` now prints if it runs in "Pull Request mode", just like it did for "CI" mode before.
* Step info logging got a complete revision, to make it more helpful, especially in case the Step fails. It now included the Step's issue tracker and source code URL infos in the log directly.
* __FIX__ : `--log-level` handling fix, the previous version had issues if the log level was set to `debug`.
