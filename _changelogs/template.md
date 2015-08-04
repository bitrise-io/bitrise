## Changes

* __BREAKING__ : change 1
* change 2


## Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -L https://github.com/bitrise-io/bitrise-cli/releases/download/{{version}}/bitrise-cli-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise-cli
```

Then:

```
chmod +x /usr/local/bin/bitrise-cli
```

The first time it's mandatory to call the `setup` after the install,
and as a best practice you should run the setup every time you install a new version of `bitrise-cli`.

Doing the setup is as easy as:

```
bitrise-cli setup
```

Once the setup finishes you have everything in place to use `bitrise-cli`.
