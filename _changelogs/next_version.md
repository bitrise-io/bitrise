## Changes

* __BREAKING__ : If trigger pattern doesn't match any expression in trigger map, workflow with the given pattern name will no longer selected.
* __BREAKING__ : Step dependency model changed. From now, dependencies are in array.
  Supported dependencie managers: brew, apt-get.
  Example:
  `
  - script:
      deps:
        brew:
        - name: cmake
        - name: git
        - name: node
        apt_get:
        - name: cmake
  `
* Improved validate command output.
* From `bitrise step-info` the --id flag removed, the first cli param used az step id. (bitrise step-info script)
* Now you can use all your environment variables (secrets level, app level, workflow level and step outputs) in run_if fields and in step inputs.
* New local setp info command added `bitrise step-info step-yml your/step/yml/pth`
* Environments got new field: IsTemplate. If IsTemplate is true, environment value, will handled with built in go template solutions.
  Exmaple:
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
  ```
* Pull Request and CI mode handling fixed

## Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/bitrise/releases/download/{{1.1.3}}/bitrise-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise
```

Then:

```
chmod +x /usr/local/bin/bitrise
```

That's all, you're ready to go!

Optionally, you can call `bitrise setup` to verify that everything what's required for `bitrise` to run
is installed and available, but if you forget to do this it'll be performed the first
time you call `bitrise run`.
