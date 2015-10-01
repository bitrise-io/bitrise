## Changes

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
* `stepman step-info` output now contains the environment inputs default_value, value_options and is_expand value.
* `stepman step-info` got new option `--step-yml`, which allows print step info about local steps.
* log improvements


## Install

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/{{0.9.17}}/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```
