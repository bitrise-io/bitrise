## Changes

* __BREAKING__ : change 1
* change 2


## Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/bitrise/releases/download/{{version}}/bitrise-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise
```

Then:

```
chmod +x /usr/local/bin/bitrise
```

That's all, you're ready to go!

Optionally, you can call `bitrise setup` to verify that everything what's required for `bitrise` to run
is installed and available, but if you forget to do this it'll be performed the first
time you call `bitrise run`.
