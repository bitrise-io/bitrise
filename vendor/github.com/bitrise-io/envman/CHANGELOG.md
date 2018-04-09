# Changelog

-----------------

## 1.1.12 (2018 Apr 09)

### Release Notes

* go dependencies update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.12/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.11 -> 1.1.12

* [10cf4bc] Krisztián  Gödrei - prepare for 1.1.12 (2018 Apr 09)
* [f182eac] Krisztián Gödrei - dependencies update (#131) (2018 Apr 09)


## 1.1.11 (2018 Mar 12)

### Release Notes

* go dependencies update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.11/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.10 -> 1.1.11

* [eef4a31] Krisztian Dobmayer - Dep update (2018 Mar 12)
* [448fdaf] Krisztian Dobmayer - Bump version to 1.1.11 (2018 Mar 12)


## 1.1.10 (2018 Feb 12)

### Release Notes

* go dependencies update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.10/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.9 -> 1.1.10

* [f066882] trapacska - Prepare for 1.1.10 (2018 Feb 12)
* [039f73d] Tamas Papik - dep-update & updated README.md (#129) (2018 Feb 12)


## 1.1.9 (2018 Jan 09)

### Release Notes

* go dependencies update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.9/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.8 -> 1.1.9

* [32b29f0] godrei - prepare for 1.1.9 (2018 Jan 09)
* [b69e031] Krisztián Gödrei - lock go-utils package (#128) (2018 Jan 09)
* [bb820f1] Krisztián Gödrei - dep update (#127) (2018 Jan 08)


## 1.1.8 (2017 Oct 09)

### Release Notes

* dependency updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.8/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.7 -> 1.1.8

* [c80708d] Krisztián Gödrei - prepare for 1.1.8 (2017 Oct 09)
* [f9d7874] Krisztián Gödrei - dep updates (#126) (2017 Oct 09)


## 1.1.7 (2017 Sep 12)

### Release Notes

* manage dependencies with [dep](https://github.com/golang/dep)
* dependency updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.7/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.6 -> 1.1.7

* [3647797] Krisztián Gödrei - prepare for 1.1.7 (2017 Sep 12)
* [ad7edf3] Krisztián Gödrei - manage dependencies with dep, dependency updates (#125) (2017 Sep 12)


## 1.1.6 (2017 Aug 07)

### Release Notes

__Meta field (`meta`) added to `EnvironmentItemOptionsModel`__

This property of the environment options is used to define extra options without creating a new [envman](https://github.com/bitrise-io/envman) release.

The __bitrise-cli__ does not use `meta` field directly, but other tools can use this property to expand the environment options.

For example the `bitrise.io` website will use the `meta` field to define if secret environment variables should be used in pull request triggered builds or not.

```
.bitrise.secrets.yml

envs:
- MY_SECRET_ENV: secret value
  opts:
    meta:
      is_expose: true
```

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.6/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.5 -> 1.1.6

* [f291011] Krisztian Godrei - prepare for 1.1.6 (2017 Aug 07)
* [1f04b73] Krisztián Gödrei - meta field added to env model (#124) (2017 Aug 07)
* [318a433] Krisztián Gödrei - fixed EnvsSerializeModel (#123) (2017 Aug 04)


## 1.1.5 (2017 Jul 10)

### Release Notes

* dependency updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.5/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.4 -> 1.1.5

* [e42d060] Krisztian Godrei - prepare for 1.1.5 (2017 Jul 10)
* [5174c25] Krisztián Gödrei - godeps update (#122) (2017 Jul 10)
* [c3aba6c] Krisztian Godrei - readme update (2017 Jul 10)
* [a9e6184] Krisztian Godrei - README update (2017 Jun 12)
* [4f0bef2] Krisztian Godrei - README: release a new version (2017 Jun 12)


## 1.1.4 (2017 Jun 12)

### Release Notes

* dependency updates
* typo fixes
* Go example updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.4/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.3 -> 1.1.4

* [5f8bb71] Krisztian Godrei - prepare for 1.1.4 (2017 Jun 12)
* [dcf6169] Krisztián Gödrei - godeps-update (#121) (2017 Jun 12)
* [e9f05f9] Viktor Benei - Go example updates (#118) (2017 May 19)
* [c3699e9] Viktor Benei - one more typo fix in README (#120) (2017 May 19)
* [1026f0c] Viktor Benei - typo fix in readme (#119) (2017 May 19)


## 1.1.3 (2017 Jan 10)

### Release Notes

* `EnvsYMLModel` model refactored to `EnvsSerializeModel` and got `Normalize` function to make the EnvsYMLModel instance json serializable, even if it was created with yml parser.
* typo fixes

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.3/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.2 -> 1.1.3

* [2c42f32] Krisztian Godrei - prepare for 1.1.3 (2017 Jan 10)
* [39332db] Krisztián Gödrei - godeps-update (#117) (2017 Jan 10)
* [0507a11] Viktor Benei - Merge pull request #116 from bitrise-io/feature/serialize-model-fix (2016 Dec 14)
* [a8b2f8f] Viktor Benei - test and doc comment for EnvsSerializeModel.Normalize (2016 Dec 14)
* [97e7667] Viktor Benei - errcheck fix (2016 Dec 14)
* [13d2bd2] Viktor Benei - EnvsYMLModel -> EnvsSerializeModel ; EnvsSerializeModel.Normalize (2016 Dec 14)
* [8c4944e] Krisztián Gödrei - Update CHANGELOG.md (2016 Nov 08)


## 1.1.2 (2016 Nov 08)

### Release Notes

* New field added to envman's environment item model: `category`.   
This new property will be used on bitrise website, to group inputs, like: `required`, `code-sign`, `xcodebuild-configs`, ...  

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.2/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.1 -> 1.1.2

* [bbac7e0] Krisztian Godrei - workflow refactor (2016 Nov 08)
* [39da611] Krisztián Gödrei - bitrise.yml update (#115) (2016 Nov 08)
* [cc31966] Krisztián Gödrei - godeps-update (#114) (2016 Nov 08)
* [9dda617] Krisztián Gödrei - add catgeory to environment model (#113) (2016 Nov 04)


## 1.1.1 (2016 Sep 13)

### Release Notes

* __NEW COMMAND:__ `envman version`   
  Prints envman's version.  
  use `--format` flag to specify output format, available options: [`json`, `yml` and `raw`]  
  use `--full` flag to print build informations like `build_number` and `commit` of the release.
* dependency updates
* unit test updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.1/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.1.0 -> 1.1.1

* [85ee7fc] Krisztián Gödrei - release config (2016 Sep 13)
* [719a45b] Krisztián Gödrei - prepare for 1.1.1 (2016 Sep 13)
* [f75917d] Krisztián Gödrei - Cli package update (#111) (2016 Sep 13)
* [92a6239] Viktor Benei - Merge pull request #110 from bitrise-io/feature/deps-update (2016 Sep 06)
* [a9776c6] Viktor Benei - godeps update (2016 Sep 06)
* [301e743] Viktor Benei - bitrise.yml update (2016 Sep 06)
* [106ffc1] Viktor Benei - gows.yml (2016 Jun 03)
* [c6d4b74] Krisztián Gödrei - Merge pull request #109 from godrei/release (2016 Apr 08)
* [f80dde4] godrei - version fix (2016 Apr 07)
* [846f23a] godrei - attach binary to release (2016 Apr 07)
* [70dc8cf] Krisztián Gödrei - Merge pull request #108 from godrei/godeps_update (2016 Apr 07)
* [cd79125] godrei - godeps update & bitrise.yml updates for go 1.6 (2016 Apr 07)
* [8046e19] Viktor Benei - Merge pull request #107 from godrei/expand_env_test (2016 Mar 16)
* [188a69a] Krisztián Gödrei - expand envs test (2016 Mar 16)
* [a99d6da] Viktor Benei - Merge pull request #106 from godrei/model_methods_update (2016 Mar 01)
* [e6de87c] Krisztián Gödrei - NewEnvJSONList instead of CreateFromJSON (2016 Mar 01)


## 1.1.0 (2015 Dec 22)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.1.0/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 1.0.0 -> 1.1.0

* [0345e3f] Viktor Benei - Merge pull request #105 from viktorbenei/master (2015 Dec 22)
* [3b330d5] Viktor Benei - changelog for version 1.1.0 (2015 Dec 22)
* [6c87fb8] Viktor Benei - godeps-update (2015 Dec 22)
* [edb2367] Viktor Benei - Dockerfile : updated Bitrise CLI version to 1.2.4 (2015 Dec 22)
* [387cc3d] Viktor Benei - Dockerfile: fix Golang version (1.5.2) instead of "1.5 latest" (2015 Dec 22)
* [efc39cb] Viktor Benei - Merge pull request #104 from viktorbenei/master (2015 Dec 22)
* [d5b3098] Viktor Benei - godeps-update (2015 Dec 22)
* [6f5c49c] Viktor Benei - Merge pull request #102 from godrei/changelog (2015 Dec 17)
* [7a1a388] Viktor Benei - Merge pull request #103 from godrei/typo (2015 Dec 17)
* [ea54069] Krisztián Gödrei - typo (2015 Dec 17)
* [34e03b4] Krisztián Gödrei - PR fixes (2015 Dec 17)
* [e650aed] Krisztián Gödrei - changelog fix (2015 Dec 16)
* [8589079] Krisztián Gödrei - version bump, changelog (2015 Dec 16)
* [099d4db] Viktor Benei - Merge pull request #100 from godrei/skip_if_empty (2015 Dec 16)
* [2ed2734] Krisztián Gödrei - flag usage updates (2015 Dec 16)
* [a105e06] Krisztián Gödrei - create changelog workflow (2015 Dec 16)
* [d59437a] Krisztián Gödrei - environment skip_if_empty property and handling (2015 Dec 16)
* [d96ed92] Krisztián Gödrei - godep update (2015 Dec 16)
* [26edf8e] Viktor Benei - Merge pull request #99 from godrei/master (2015 Dec 15)
* [33ddda6] Krisztián Gödrei - godeps update (2015 Dec 15)
* [9b9e8a5] Krisztián Gödrei - skip if empty field (2015 Dec 15)


## 1.0.0 (2015 Oct 31)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/1.0.0/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.10 -> 1.0.0

* [88d919d] Viktor Benei - Merge pull request #97 from viktorbenei/master (2015 Oct 31)
* [a5fd047] Viktor Benei - changelog v1.0.0 (2015 Oct 31)
* [bbbaea6] Viktor Benei - Merge pull request #96 from viktorbenei/master (2015 Oct 31)
* [c12dc12] Viktor Benei - v1.0.0 (2015 Oct 31)
* [291c3b9] Viktor Benei - Merge pull request #95 from viktorbenei/master (2015 Oct 31)
* [af12dcb] Viktor Benei - fixing a couple of issues with the new Configs handling and Max Env Size handling (2015 Oct 31)
* [c1e0918] Viktor Benei - bitrise.yml : test_and_install workflow fix, to actually call go install (2015 Oct 31)
* [ca1bc28] Viktor Benei - godeps-update (2015 Oct 31)
* [a5b6120] Viktor Benei - Merge branch 'gkiki90-cmd_line_length' (2015 Oct 31)
* [430ef8f] Viktor Benei - Merge branch 'cmd_line_length' of https://github.com/gkiki90/envman into gkiki90-cmd_line_length (2015 Oct 31)
* [2426933] Viktor Benei - Merge pull request #93 from gkiki90/log_fix (2015 Oct 31)
* [3d4b86b] Krisztian Godrei - PR fix (2015 Oct 27)
* [639b512] Krisztian Godrei - PR fix (2015 Oct 27)
* [9128136] Krisztian Godrei - PR fix (2015 Oct 27)
* [a477d0f] Krisztian Godrei - changelog (2015 Oct 27)
* [7086e58] Krisztian Godrei - env length fix (2015 Oct 27)
* [efcbb40] Krisztian Godrei - PR fix (2015 Oct 27)
* [e40a476] Krisztian Godrei - changelog (2015 Oct 26)
* [385f496] Krisztian Godrei - exit code fix (2015 Oct 26)
* [48688c4] Krisztian Godrei - upcoming.md fix (2015 Oct 26)
* [55131ad] Krisztian Godrei - removed unnecessary log (2015 Oct 26)
* [85c7d0f] Viktor Benei - Merge pull request #92 from viktorbenei/master (2015 Oct 02)
* [2e18704] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Oct 02)
* [7005d48] Viktor Benei - `DefaultIsTemplate` - typo fix (missing `e`) (2015 Oct 02)


## 0.9.10 (2015 Oct 02)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.10/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.9 -> 0.9.10

* [22b6ba3] Viktor Benei - Merge pull request #91 from viktorbenei/master (2015 Oct 02)
* [07e36ba] Viktor Benei - v0.9.10 changelog (2015 Oct 02)
* [cf76633] Viktor Benei - v0.9.10 (2015 Oct 02)
* [56798e0] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Oct 02)
* [826ecb3] Viktor Benei - godeps-update (2015 Oct 02)
* [263f8d3] Viktor Benei - Merge pull request #90 from gkiki90/changelog (2015 Oct 01)
* [443801e] Krisztian Goedrei - changelog (2015 Oct 01)
* [8c81521] Viktor Benei - Merge pull request #89 from gkiki90/parse_bool (2015 Oct 01)
* [9980e7b] Krisztian Goedrei - parse bools (2015 Oct 01)
* [9d74a12] Viktor Benei - Merge pull request #88 from gkiki90/constructor (2015 Sep 30)
* [74f2133] Krisztian Goedrei - json constructor (2015 Sep 30)
* [8098976] Viktor Benei - Merge pull request #87 from gkiki90/cast_value (2015 Sep 29)
* [1658849] Krisztian Goedrei - value, value_options cast to string (2015 Sep 28)
* [c58637a] Viktor Benei - Merge pull request #86 from gkiki90/input_template (2015 Sep 25)
* [d4e83c1] Krisztian Goedrei - godep (2015 Sep 25)
* [f7fe2a2] Krisztian Goedrei - require in tests (2015 Sep 25)
* [941b8e2] Krisztian Goedrei - isTemplate option,  require in test (2015 Sep 25)


## 0.9.9 (2015 Sep 21)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.9/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.8 -> 0.9.9

* [dd37e94] Viktor Benei - Merge pull request #85 from viktorbenei/master (2015 Sep 21)
* [da75b63] Viktor Benei - finalizing 0.9.9 (2015 Sep 21)
* [306e536] Viktor Benei - Merge pull request #83 from gkiki90/change_log (2015 Sep 21)
* [49c549b] Viktor Benei - Merge pull request #84 from gkiki90/ci (2015 Sep 21)
* [4c5d7e7] Krisztian Goedrei - ci fix (2015 Sep 21)
* [b5a9c41] Krisztian Goedrei - change log (2015 Sep 21)
* [dd4ac87] Viktor Benei - Merge pull request #82 from gkiki90/print (2015 Sep 21)
* [b5f3527] Krisztian Goedrei - print fix (2015 Sep 21)
* [f697b5f] Viktor Benei - Merge pull request #81 from gkiki90/envman_print (2015 Sep 21)
* [f220884] Krisztian Goedrei - godeps (2015 Sep 21)
* [97931a5] Krisztian Goedrei - print with format and expand options (2015 Sep 21)
* [2ff0b09] Viktor Benei - Merge pull request #80 from viktorbenei/master (2015 Sep 18)
* [dc40683] Viktor Benei - changelog install `curl` fix (2015 Sep 16)
* [f0275dc] Viktor Benei - Merge pull request #79 from viktorbenei/master (2015 Sep 16)
* [748134b] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Sep 16)
* [3f27e51] Viktor Benei - `errcheck` script revision in `bitrise.yml` (2015 Sep 16)
* [b19a192] Viktor Benei - Merge pull request #78 from viktorbenei/master (2015 Sep 16)
* [2dda223] Viktor Benei - godeps-update (2015 Sep 16)
* [1242452] Viktor Benei - Merge pull request #77 from gkiki90/ci (2015 Sep 15)
* [caa3913] Krisztian Goedrei - ci fix (2015 Sep 15)
* [db7c8ce] Krisztian Goedrei - code style (2015 Sep 14)
* [590a720] Krisztian Goedrei - fix (2015 Sep 14)
* [3cd85bb] Krisztian Goedrei - new ci (2015 Sep 14)
* [86840cc] Viktor Benei - Merge pull request #76 from viktorbenei/master (2015 Sep 08)
* [10d74ef] Viktor Benei - full godeps-update (2015 Sep 08)
* [5532e41] Viktor Benei - Merge pull request #75 from gkiki90/readme (2015 Sep 08)
* [9694531] Krisztian Goedrei - update (2015 Sep 08)
* [9754f3e] Viktor Benei - Merge pull request #74 from viktorbenei/master (2015 Sep 07)
* [34418e0] Viktor Benei - Docker file : unified with other projects (`stepman`) (2015 Sep 07)
* [b0b3dc5] Viktor Benei - godeps-update fix, to include every package (2015 Sep 07)
* [7f17fd2] Viktor Benei - Merge pull request #73 from viktorbenei/master (2015 Sep 07)
* [f1cd92f] Viktor Benei - renamed 'install bitrise CLI' script (2015 Sep 07)
* [eae7259] Viktor Benei - start of v0.9.9 (2015 Sep 07)


## 0.9.8 (2015 Sep 07)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.8/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.7 -> 0.9.8

* [b49ba68] Viktor Benei - Merge pull request #72 from viktorbenei/master (2015 Sep 07)
* [ad02218] Viktor Benei - docker related updates & changelog for 0.9.8 (2015 Sep 07)
* [45ccf59] Viktor Benei - goddess update (2015 Sep 07)
* [a629239] Viktor Benei - Merge pull request #71 from viktorbenei/master (2015 Sep 02)
* [b46109d] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Sep 02)
* [3f8f285] Viktor Benei - `golint` now fails in CI if finds issues (2015 Sep 02)
* [93cb295] Viktor Benei - Merge pull request #70 from gkiki90/model_fix (2015 Sep 02)
* [bd3e8af] Krisztian Goedrei - [282265e] format version [eaf5c0c] model fix (2015 Sep 02)
* [8aa9bcb] Viktor Benei - Merge pull request #69 from viktorbenei/master (2015 Aug 31)
* [9e925d4] Viktor Benei - start of v0.9.8 (2015 Aug 31)


## 0.9.7 (2015 Aug 31)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.7/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.6 -> 0.9.7

* [3f4b5be] Viktor Benei - Merge pull request #68 from viktorbenei/master (2015 Aug 31)
* [8cb4fee] Viktor Benei - changelog v0.9.7 (2015 Aug 31)
* [91d933c] Viktor Benei - Merge pull request #67 from viktorbenei/master (2015 Aug 31)
* [ec4c9a3] Viktor Benei - godeps-update (2015 Aug 31)
* [591f7fb] Viktor Benei - Merge pull request #66 from viktorbenei/master (2015 Aug 28)
* [8518f29] Viktor Benei - godeps-update (2015 Aug 28)
* [1c6ddd6] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Aug 28)
* [a45b945] Viktor Benei - Merge pull request #65 from gkiki90/ci (2015 Aug 27)
* [24e20eb] Krisztian Goedrei - ci (2015 Aug 27)
* [4100017] Viktor Benei - Merge pull request #64 from gkiki90/master (2015 Aug 27)
* [6f5bb1b] Krisztian Goedrei - force color log (2015 Aug 27)
* [509efb2] Viktor Benei - Merge pull request #63 from gkiki90/init_fixes (2015 Aug 26)
* [a25959b] Viktor Benei - godeps-update (2015 Aug 26)
* [ae9b09d] Krisztian Goedrei - init fix (2015 Aug 25)
* [f4d2c83] Viktor Benei - Merge pull request #62 from bazscsa/master (2015 Aug 24)
* [5b29e2c] Tamás Bazsonyi - Show envman releases in the Install guide section (2015 Aug 24)
* [20f921c] Viktor Benei - Merge pull request #61 from viktorbenei/master (2015 Aug 24)
* [8f8deec] Viktor Benei - start of v0.9.7 (2015 Aug 24)
* [8fc5738] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/envman (2015 Aug 24)
* [ce9a927] Viktor Benei - bitrise.yml update : for the new `slack` step (2015 Aug 24)


## 0.9.6 (2015 Aug 24)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.6/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.5 -> 0.9.6

* [8cd7a5e] Viktor Benei - Merge pull request #60 from viktorbenei/master (2015 Aug 24)
* [1badd24] Viktor Benei - changelog v0.9.6 (2015 Aug 24)
* [8ba15c4] Viktor Benei - Merge pull request #59 from gkiki90/title_desc_summary (2015 Aug 19)
* [5770aef] Krisztian Goedrei - Summary field (2015 Aug 18)
* [7658fe6] Viktor Benei - Merge pull request #58 from viktorbenei/master (2015 Aug 17)
* [d02293d] Viktor Benei - FIX: env item option model casting should work for map[string]interface, not just for map[interface]interface (2015 Aug 17)
* [4c311c1] Viktor Benei - typo fix (2015 Aug 14)
* [fa00328] Viktor Benei - Update README.md (2015 Aug 14)
* [4f6a847] Viktor Benei - Merge pull request #57 from viktorbenei/master (2015 Aug 14)
* [8e5d433] Viktor Benei - start of v0.9.6 (2015 Aug 14)


## 0.9.5 (2015 Aug 14)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.5/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.4 -> 0.9.5

* [24629b6] Viktor Benei - Merge pull request #56 from viktorbenei/master (2015 Aug 14)
* [a1814bd] Viktor Benei - changelog (2015 Aug 14)
* [4fdeb28] Viktor Benei - Merge pull request #55 from gkiki90/go-utils_cli-fix (2015 Aug 13)
* [cb55d0d] Krisztian Goedrei -  go-util, cli fixes (2015 Aug 13)
* [17713a6] Viktor Benei - Merge pull request #54 from viktorbenei/master (2015 Aug 12)
* [a844b22] Viktor Benei - start of v0.9.5 (2015 Aug 12)


## 0.9.4 (2015 Aug 12)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.4/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.3 -> 0.9.4

* [83c9ea2] Viktor Benei - Merge pull request #53 from viktorbenei/master (2015 Aug 12)
* [08fe405] Viktor Benei - godeps-update (2015 Aug 12)
* [1dbd994] Viktor Benei - Merge pull request #52 from viktorbenei/master (2015 Aug 11)
* [f513cb3] Viktor Benei - godeps-update (2015 Aug 11)
* [7c854de] Viktor Benei - go-utils/pointers revision / migration (2015 Aug 11)
* [60e0505] Viktor Benei - Merge pull request #51 from viktorbenei/master (2015 Aug 11)
* [b4509ae] Viktor Benei - updated the pathutil package to be used from the new go-utils repo (2015 Aug 11)
* [2565e5e] Viktor Benei - Merge pull request #50 from gkiki90/go-util_pointers (2015 Aug 10)
* [892a957] Krisztian Goedrei - pointers (2015 Aug 10)
* [6231b9c] Viktor Benei - Merge pull request #49 from gkiki90/ptr (2015 Aug 10)
* [c0f2c46] Krisztian Goedrei - godep update (+1 squashed commit) Squashed commits: [8e99b8e] ptr with go-utils (2015 Aug 10)
* [b488588] Viktor Benei - Merge pull request #48 from gkiki90/model_fix (2015 Aug 10)
* [3016845] Krisztian Goedrei - test fixes (+2 squashed commits) Squashed commits: [26ab05d] var to const [784b3d6] fill missing defaults fix (2015 Aug 10)
* [3bc0732] Viktor Benei - Merge pull request #47 from viktorbenei/master (2015 Aug 08)
* [1bb542d] Viktor Benei - goddess update (2015 Aug 08)
* [9c38f04] Viktor Benei - Merge pull request #46 from viktorbenei/master (2015 Aug 08)
* [721ba33] Viktor Benei - updated default-step-lib-source in bitrise.yml (2015 Aug 08)
* [f45f360] Viktor Benei - Merge pull request #45 from viktorbenei/master (2015 Aug 05)
* [10fd6b0] Viktor Benei - start of v0.9.4 (2015 Aug 05)


## 0.9.3 (2015 Aug 05)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.3/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.2 -> 0.9.3

* [069394d] Viktor Benei - Merge pull request #44 from gkiki90/master (2015 Aug 05)
* [d1d00f5] Krisztian Goedrei - changelog (2015 Aug 05)
* [396133c] Viktor Benei - Merge pull request #43 from gkiki90/master (2015 Aug 05)
* [9eab5c7] Krisztian Goedrei - flag fixes (2015 Aug 05)
* [0940a21] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/envman (2015 Aug 05)
* [c6127d9] Krisztian Goedrei - bool flag fix (2015 Aug 05)
* [9ff8a84] Viktor Benei - Merge pull request #42 from viktorbenei/master (2015 Aug 05)
* [f31d70d] Viktor Benei - changelog template fix (2015 Aug 04)
* [8f0139f] Viktor Benei - Merge pull request #41 from viktorbenei/master (2015 Aug 03)
* [c981ba5] Viktor Benei - Merge pull request #40 from gkiki90/master (2015 Aug 03)
* [9b8aac1] Viktor Benei - start of v0.9.3 (2015 Aug 03)
* [7a6fea9] Krisztian Goedrei - godep-update (2015 Aug 03)
* [3b0b628] Viktor Benei - Install instructions now points to /releases (2015 Aug 03)


## 0.9.2 (2015 Aug 03)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/envman/releases/download/0.9.2/envman-$(uname -s)-$(uname -m) > /usr/local/bin/envman
```

Then:

```
chmod +x /usr/local/bin/envman
```

That's all, you're ready to call `envman`!

### Release Commits - 0.9.1 -> 0.9.2

* [c03579c] Viktor Benei - Merge pull request #39 from gkiki90/release (2015 Aug 03)
* [3c07d19] Krisztian Goedrei - changelog handling (2015 Aug 03)
* [847a827] Krisztian Goedrei - bitrise.yml for release (2015 Aug 03)
* [315c956] Viktor Benei - Merge pull request #38 from gkiki90/tests (2015 Aug 03)
* [817eab8] Krisztian Goedrei - test fixes (2015 Aug 03)
* [eb736e9] Krisztian Goedrei - code style (2015 Aug 03)
* [9fa878c] Krisztian Goedrei - errcheck (2015 Aug 03)
* [fe13af3] Krisztian Goedrei - runCmd fix, test fixes (2015 Aug 03)
* [50cb049] Krisztian Goedrei - code style (2015 Aug 03)
* [1e5ed59] Krisztian Goedrei - code style (2015 Aug 03)
* [77e9d37] Krisztian Goedrei - PR fixes (2015 Aug 03)
* [2c0237e] Krisztian Goedrei - code cleaning (2015 Jul 30)
* [76b0679] Krisztian Goedrei - envman tests, refactor, code style (2015 Jul 30)
* [dc2189c] Krisztian Goedrei - models_methods_test (2015 Jul 30)
* [2386964] Krisztian Goedrei - merge with stash (2015 Jul 30)
* [2a36e84] Viktor Benei - Merge pull request #37 from viktorbenei/master (2015 Jul 29)
* [a0a5b49] Viktor Benei - include the invalid env if GetKeyValuePair fails because of more than 2 fields found (2015 Jul 29)
* [387b3f0] Krisztian Goedrei - test in progress (2015 Jul 29)
* [7a807fd] Krisztian Goedrei - test prefill (2015 Jul 29)
* [2447c35] Viktor Benei - Merge pull request #36 from gkiki90/new_models (2015 Jul 29)
* [e719d0a] Krisztian Goedrei - PR fixes (2015 Jul 29)
* [a739054] Krisztian Goedrei - fix (2015 Jul 29)
* [23f74de] Krisztian Goedrei - new model handling (2015 Jul 29)
* [66dc1d4] Krisztian Goedrei - new env model (2015 Jul 29)
* [dbd3a06] Viktor Benei - install format change (2015 Jul 22)
* [09caaef] Viktor Benei - Install instructions (2015 Jul 17)


-----------------

Generated at: 2018 Apr 09
