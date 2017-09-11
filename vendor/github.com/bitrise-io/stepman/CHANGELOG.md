# Changelog

-----------------

## 0.9.33 (2017 Aug 07)

### Release Notes

* go dependencies update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.33/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.32 -> 0.9.33

* [1cb040a] Krisztian Godrei - prepare for 0.9.33 (2017 Aug 07)
* [05d6b20] Krisztián Gödrei - godeps-update (#235) (2017 Aug 07)


## 0.9.32 (2017 Jul 10)

### Release Notes

* git step's default branch is `master`, instead of the repository's default branch.

This means: if you use a step from it's git source, and do not specify the repo's branch to use:

```
workflows:
  primary:
    steps:
    - git::https://github.com/bitrise-community/steps-ionic-archive.git: 
```

the master branch will be cloned.

* dependency updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.32/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.31 -> 0.9.32

* [ec41577] Krisztian Godrei - prepare for 0.9.32 (2017 Jul 10)
* [171cae1] Krisztián Gödrei - godeps update (#234) (2017 Jul 10)
* [58796bd] Krisztián Gödrei - integration test update, git step default branch is master (#233) (2017 Jul 04)
* [7e77ca8] Krisztian Godrei - README: Release a new version (2017 Jun 12)


## 0.9.31 (2017 Jun 12)

### Release Notes

* godeps-update

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.31/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.30 -> 0.9.31

* [f53b5da] Krisztian Godrei - prepare for 0.9.31 (2017 Jun 12)
* [5a930c6] Krisztián Gödrei - godeps-update (#232) (2017 Jun 12)


## 0.9.30 (2017 Apr 10)

### Release Notes

* `step-info` command fix in case of git type step: if step version not specified, stepman only does a git clone, instead of force setting `master` branch in clone.
* better error messages in `step-info` command
* logging updates

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.30/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.29 -> 0.9.30

* [a9cafb8] Krisztian Godrei - prepare for 0.9.30 (2017 Apr 10)
* [a53b457] Krisztián Gödrei - godeps update (#231) (2017 Apr 10)
* [d723791] Krisztián Gödrei - do not force master if branch not specified (#230) (2017 Apr 10)
* [927e227] Tamas Papik - Updated step-info logging (#229) (2017 Apr 10)


## 0.9.29 (2017 Mar 13)

### Release Notes

* Every git command bundled in retry block, to avoid github networking issues
* Log improvements, for better error messages and logs

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.29/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.28 -> 0.9.29

* [063e8d2] Krisztian Godrei - prepare for 0.9.29 (2017 Mar 13)
* [91afbe9] Krisztián Gödrei - godeps update (#228) (2017 Mar 13)
* [c091444] Krisztián Gödrei - warn fix (#227) (2017 Feb 24)
* [41dd088] Tamas Papik - Git commands bundled in retry (#226) (2017 Feb 20)
* [e31a6b9] Tamas Papik - Unused function removed (#225) (2017 Feb 16)
* [0fd591f] Krisztián Gödrei - Update CHANGELOG.md (2017 Feb 14)
* [836e01c] Krisztian Godrei - changelog update (2017 Feb 14)


## 0.9.28 (2017 Feb 14)

### Release Notes

__BREAKING__ : `step-info` command revision: 

We added local and git step support to this command and also updated its log, to provide all neccessary infos about every type of the supported steps.

You can specify the step's source with the new `--library` flag. As a value you can provide: 

- `STEPLI_URI` - the git uri of the step library
- `path` - specifies local step
- `git` - if you want to use a step from its git source

With `--id` flag, you can specify the unique identifier of the step in its collection:

- in case of __step library step__: the unique identifier in the library
- in case of __local step__: the local path of the step directory
- in case of __git step__: the git uri of the step reporitory

`--version` flag:

- in case of __steplib step__: the step version in the steplib
- in case of __local step__: _not used_
- in case of __git step__: git tag or branch

You can define the __output format__ of the command by passing `--format FORMAT` flag. 

Format can be either `raw` (default):

```
$ stepman step-info --library https://github.com/bitrise-io/bitrise-steplib.git --id script --version 1.1.1 --format raw

Library: https://github.com/bitrise-io/bitrise-steplib.git
ID: script
Version: 1.1.1
LatestVersion: 1.1.3
Definition:

[step.yml content]
```

or `json` to use the command's output by other tools:

```
$ stepman step-info --library https://github.com/bitrise-io/bitrise-steplib.git --id script --version 1.1.1 --format json

{
   "library":"https://github.com/bitrise-io/bitrise-steplib.git",
   "id":"script",
   "version":"1.1.1",
   "latest_version":"1.1.3",
   "info":{

   },
   "step":{

     [serialized step model]

   },
   "definition_pth":"$HOME/.stepman/step_collections/1487001505/collection/steps/script/1.1.1/step.yml"
}
```

__Examples:__

Get info about a step from the step library:

`stepman step-info --library https://github.com/bitrise-io/bitrise-steplib.git --id script --version 1.1.1`

Get info about a local step:

`stepman step-info --library path --id /PATH/TO/THE/STEP/DIRECTORY`

Get step info about a step, defined by its git repository uri:

`stepman step-info --library git --id https://github.com/bitrise-io/steps-script.git --version master`

Command flag changes:

- `--collection` is deprecated, use `--library` instead
- `--short` is deprecated and no longer used
- `--step-yml` is deprecated, use `--library path` and `--id PATH_TO_YOUR_STEP_DIR` instead

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.28/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.27 -> 0.9.28

* [ce0083d] Krisztian Godrei - release workflow updates & preare for 0.9.28 (2017 Feb 14)
* [3355572] Krisztián Gödrei - step-info revision (#224) (2017 Feb 13)


## 0.9.27 (2017 Jan 25)

### Release Notes

- `stepman collections` command now prints the collection's spec.json path as well:

```
https://github.com/bitrise-io/bitrise-steplib.git
  spec_path: $HOME/.stepman/step_collections/1485356810/spec/spec.json
```

This update allows the [Workflow Editor](https://github.com/bitrise-io/bitrise-workflow-editor) to use local steplib spec through stepman, instead of custom logic.

- Use the new command package ([go-utils/command](https://github.com/bitrise-io/go-utils/tree/master/command)) instead of previous version ([go-utils/cmdex](https://github.com/bitrise-io/go-utils/pull/44/files))


### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.27/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.26 -> 0.9.27

* [c03c13a] Krisztian Godrei - godeps update (2017 Jan 25)
* [e649be4] Krisztian Godrei - prepare for 0.9.27 (2017 Jan 25)
* [876ce6f] Krisztián Gödrei - print SteplibInfoModel in collections command (#223) (2017 Jan 16)


## 0.9.26 (2016 Dec 13)

### Release Notes

* StepModel got a new property: `Timeout`. This new property prepares a feature step timeout handling. 
* StepModel json and yml representation now ommits empty `Source` and `Deps` properties, intsead of printing empty struct for this values. 
* `step-list` command revision, to easily get summary of steps in the specified steplib. 

steplis item looks like:

```
 * STEP_TITLE
   ID: STEP_ID
   Latest Version: LATEST_VERSION
   Summary: STEP_SUMMARY
``` 

for example:

```
 * Sign APK
   ID: sign-apk
   Latest Version: 1.1.1
   Summary: Sign APK
```

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.26/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.25 -> 0.9.26

* [1318645] Krisztian Godrei - prepare for 0.9.26 (2016 Dec 13)
* [df0ff3d] Viktor Benei - step info and step list revs (#222) (2016 Dec 12)
* [80f3ca5] Krisztián Gödrei - omitt source and deps properties if empty (#221) (2016 Nov 29)
* [fd38d19] Krisztián Gödrei - add timeout to step model (#220) (2016 Nov 24)


## 0.9.25 (2016 Oct 14)

### Release Notes

* `stepman share` command fix: in version 0.9.24 stepman created a branch - for sharing a new step - with name: `STEP_ID` and later tried to push the steplib changes on branch: `STEP_ID-STEP_VERSION`, which branch does not exist.  
This release contains a quick fix for stepman sharing, the final share branch layout is: `STEP_ID-STEP_VERSION`

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.25/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.24 -> 0.9.25

* [a800ae9] Krisztian Godrei - prepare for 0.9.25 (2016 Oct 14)
* [594670c] Krisztián Gödrei - Share fix (#218) (2016 Oct 14)


## 0.9.24 (2016 Oct 11)

### Release Notes

* step version added to step share branch. New share branch layout: `STEP_ID-STEP_VERSION`.
* some error message fixes 

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.24/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.23 -> 0.9.24

* [c7090f1] Krisztian Godrei - prepare for 0.9.24 (2016 Oct 11)
* [f3461f0] Krisztián Gödrei - error log fix (#217) (2016 Oct 06)
* [43bd2bb] Krisztián Gödrei - add step version to share branch name (#216) (2016 Oct 06)
* [ab2eb60] Viktor Benei - minor duplication fix (2016 Sep 16)


## 0.9.23 (2016 Sep 13)

### Release Notes

* __Toolkit support:__  Currently available toolkits: `bash` and `go`.
  * If a step utilizes a Toolkit it does not have to provide a bash entry file (`step.sh`) anymore (except using bash toolkit).
  * Using the toolkit can also provide __performance benefits__, as it does automatic binary caching - which means that a given version of the step will only be compiled the first time, subsequent execution of the same version will use the compiled binary of the step.  
  See more about Toolkit on bitrise cli's 1.4.0 [release page](https://github.com/bitrise-io/bitrise/releases/tag/1.4.0).
* __Dependecy models got new property:__ `bin_name`  
  bin_name is the binary's name, if it doesn't match the package's name.  
  E.g. in case of "AWS CLI" the package is `awscli` and the binary is `aws`.  
  If BinName is empty Name will be used as BinName too.
* Every __networking__ command uses __retry logic.__ 
* Better error messages.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.23/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.22 -> 0.9.23

* [04e670e] Krisztián Gödrei - Merge branch 'master' of github.com:bitrise-io/stepman (2016 Sep 13)
* [578ad1f] Krisztián Gödrei - Godep update (#214) (2016 Sep 13)
* [61ef1e5] Krisztián Gödrei - prepare for 0.9.23 (2016 Sep 13)
* [ec33a9d] Viktor Benei - Merge pull request #213 from bitrise-io/feature/deps-bin-name (2016 Sep 11)
* [ff5807b] Viktor Benei - deps GetBinaryName (2016 Sep 11)
* [aad8c23] Viktor Benei - Deps extended with BinName (2016 Sep 11)
* [5ee555b] Viktor Benei - Merge pull request #212 from bitrise-io/feature/retrying-log (2016 Sep 09)
* [ce5d89c] Viktor Benei - just a heads-up/debug log (2016 Sep 09)
* [c8d21da] Viktor Benei - Merge pull request #211 from bitrise-io/feature/audit-step-retry (2016 Sep 06)
* [ceee620] Viktor Benei - better error messages for step model audit (2016 Sep 06)
* [6e4d306] Viktor Benei - steplib audit : retry step version git clones (2016 Sep 06)
* [b959724] Viktor Benei - Merge pull request #210 from bitrise-io/feature/retries (2016 Sep 06)
* [fd38bc5] Viktor Benei - retry step download and StepLib update (2016 Sep 06)
* [cbe9f2d] Viktor Benei - Merge pull request #209 from bitrise-io/feature/deps-update (2016 Sep 06)
* [665c864] Viktor Benei - godeps update (2016 Sep 06)
* [f88d1ca] Viktor Benei - bitrise.yml minor update (2016 Sep 06)
* [574714f] Viktor Benei - Merge pull request #208 from bitrise-io/feature/step-model-toolkits (2016 Sep 06)
* [3eebcef] Viktor Benei - empty serialize example/test (2016 Sep 06)
* [40fe06c] Viktor Benei - test fix (2016 Sep 06)
* [7a881cf] Viktor Benei - toolkits - pointers (2016 Sep 06)
* [3a82567] Viktor Benei - Merge pull request #207 from bitrise-io/feature/step-model-toolkits (2016 Sep 06)
* [09eaac4] Viktor Benei - Step models and property for toolkits (2016 Sep 06)
* [4927863] Krisztián Gödrei - Merge pull request #204 from bitrise-io/viktorbenei-patch-1 (2016 Aug 11)
* [8a6da10] Viktor Benei - Log - path ref fix (2016 Aug 11)
* [97db84b] Krisztián Gödrei - Merge pull request #203 from godrei/master (2016 Jul 19)
* [d512bad] Krisztián Gödrei - changelog (2016 Jul 19)


## 0.9.22 (2016 Jul 19)

### Release Notes

* Fixed local steplib handling & integration tests.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.22/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.21 -> 0.9.22

* [2e6dec1] Krisztián Gödrei - prepare for 0.9.22 (2016 Jul 19)
* [f49f77f] Krisztián Gödrei - Merge pull request #202 from godrei/local_steplib (2016 Jul 19)
* [086baf8] Krisztián Gödrei - fixed cleanup dangling route (2016 Jul 19)
* [31b4e43] Krisztián Gödrei - PR fix (2016 Jul 19)
* [d4e345f] Krisztián Gödrei - steplib fix & tests (2016 Jul 19)
* [b2a4a26] Krisztián Gödrei - Merge pull request #200 from godrei/master (2016 Jul 12)
* [e4cf91b] Krisztián Gödrei - changelog (2016 Jul 12)


## 0.9.21 (2016 Jul 12)

### Release Notes

* Previous version (0.9.20) returned with exit code 0, even if command failed. This version fixes this issue and includes integration tests to catch this in automated tests in the future.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.21/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.20 -> 0.9.21

* [86b6172] Krisztián Gödrei - prepare for 0.9.21 (2016 Jul 12)
* [4a24916] Krisztián Gödrei - Merge pull request #199 from godrei/exit_status_fix (2016 Jul 12)
* [d8cca3a] Krisztián Gödrei - exist status test (2016 Jul 12)
* [40c8517] Krisztián Gödrei - exist status fix (2016 Jul 12)
* [92830ff] Krisztián Gödrei - Merge branch 'master' of github.com:bitrise-io/stepman (2016 Jul 12)
* [d34225e] Krisztián Gödrei - changelog update (2016 Jul 12)
* [a6b4a3f] Krisztián Gödrei - Merge pull request #198 from godrei/master (2016 Jul 12)


## 0.9.20 (2016 Jul 12)

### Release Notes

* __BREAKING__ : every command's short version has been removed.
* __BREAKING__ : `step-info` command's `collection` flag is required.
* __NEW COMMAND__ : `export-spec` - Export the generated StepLib spec, use `export-type` flag to specify the export type.  
  Export type options:

  - `full` : exports the full StepLib spec
  - `latest` : exported spec only contains steps with latest versions
  - `minimal` : exported spec's steps field only contains the step-ids
* Improved logging in `stepman update` command

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.20/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.19 -> 0.9.20

* [7f0878f] Krisztián Gödrei - prepare for 0.9.20 (2016 Jul 12)
* [ede2058] Krisztián Gödrei - Merge pull request #197 from godrei/spec (2016 Jul 12)
* [10f144b] Krisztián Gödrei - specify output path (2016 Jul 12)
* [813af05] Krisztián Gödrei - export StepLib spec (2016 Jul 12)
* [8269283] Krisztián Gödrei - Merge pull request #196 from godrei/step_info (2016 Jul 11)
* [99e557d] Krisztián Gödrei - Merge pull request #195 from godrei/short_commands (2016 Jul 11)
* [54d8d5b] Krisztián Gödrei - step-info requires StepLib (2016 Jul 11)
* [98acc03] Krisztián Gödrei - remove short commands,  delete log fix (2016 Jul 11)
* [c96f79d] Krisztián Gödrei - Merge pull request #194 from godrei/update_fix (2016 Jul 08)
* [972e1e2] Krisztián Gödrei - return with error (2016 Jul 08)
* [dc13f84] Krisztián Gödrei - steplib update logging fix (2016 Jul 08)
* [d432605] Krisztián Gödrei - Merge pull request #193 from godrei/deprecation_fix (2016 Jul 08)
* [83fdfd1] Krisztián Gödrei - deprecation fix (2016 Jul 08)
* [8306a8e] Krisztián Gödrei - Merge pull request #192 from godrei/master (2016 Jul 08)
* [145a034] Krisztián Gödrei - gows godep save (2016 Jul 08)
* [8c13bf2] Krisztián Gödrei - Merged branch master into master (2016 Jul 08)
* [7323eeb] Viktor Benei - LICENSE (MIT) (2016 Jun 03)
* [400b51a] Viktor Benei - gows.yml (2016 Jun 03)
* [f392873] Viktor Benei - Merge pull request #185 from godrei/goinp_fix (2016 May 23)
* [3be5c7b] Krisztián Gödrei - godep update (2016 May 23)
* [1084a3d] Krisztián Gödrei - goinp fix (2016 May 23)
* [294c5eb] Krisztián Gödrei - Merge pull request #184 from godrei/setup_fix (2016 May 11)
* [008dbc8] Krisztián Gödrei - setup review (2016 May 11)
* [d822310] Krisztián Gödrei - Merge pull request #183 from godrei/master (2016 May 09)
* [623e925] Krisztián Gödrei - changelog (2016 May 09)


## 0.9.19 (2016 May 09)

### Release Notes

* step-template link fix
* minor bug fixes and improvements

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.19/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.18 -> 0.9.19

* [6f06c1b] Viktor Benei - step-template link fix (2016 May 06)
* [6d4c12e] Krisztián Gödrei - Merge pull request #180 from godrei/version_cmd (2016 Apr 07)
* [f8d89af] godrei - version cmd (2016 Apr 07)
* [ba10ca9] Krisztián Gödrei - Merge pull request #179 from godrei/validation_fix (2016 Apr 05)
* [31b9f40] godrei - validate updates (2016 Apr 05)
* [0478153] Krisztián Gödrei - Merge pull request #178 from godrei/download_step_fix (2016 Apr 05)
* [736a3d9] godrei - downoad setp fix (2016 Apr 05)
* [e15f5e2] Krisztián Gödrei - Merge pull request #177 from godrei/feature/release (2016 Apr 05)
* [24159a5] godrei - PR fix (2016 Apr 05)
* [76cf1c1] godrei - release config, changelog update (2016 Apr 05)
* [319745a] godrei - in progress (2016 Apr 05)
* [240a642] Viktor Benei - Merge pull request #176 from godrei/feature/godep_update (2016 Apr 05)
* [8288a74] godrei - godeps update (2016 Apr 05)
* [8d62410] Viktor Benei - Merge pull request #173 from bitrise-io/fix/validate-collection-aliases (2016 Feb 09)
* [bf22a75] vasarhelyia - Validate step collection routes (2016 Feb 01)


## 0.9.18 (2015 Dec 22)

### Release Notes

* Step ID must conform to [a-z0-9-] regexp
* Typo fixes
* Logging revisions

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.18/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.17 -> 0.9.18

* [284f156] Viktor Benei - Merge pull request #172 from viktorbenei/master (2015 Dec 22)
* [404a02a] Viktor Benei - changelog for 0.9.18 (2015 Dec 22)
* [366ce82] Viktor Benei - moved upcoming changes from 0.9.18.md to upcoming.md (2015 Dec 22)
* [93cdcbd] Viktor Benei - Merge pull request #171 from viktorbenei/master (2015 Dec 22)
* [c92e17c] Viktor Benei - version bump: 0.9.18 (2015 Dec 22)
* [e0d8112] Viktor Benei - Merge pull request #170 from viktorbenei/master (2015 Dec 22)
* [9feb7a8] Viktor Benei - Dockerfile : fix Go version (1.5.2) & update to Bitrise CLI 1.2.4 (2015 Dec 22)
* [db278aa] Viktor Benei - godeps-update (2015 Dec 22)
* [3d2293b] Viktor Benei - Merge pull request #168 from godrei/step_info_fix (2015 Dec 22)
* [64d6c5e] Viktor Benei - Merge pull request #169 from godrei/master (2015 Dec 22)
* [84b8550] Krisztián Gödrei - create changelog (2015 Dec 22)
* [952806c] Krisztián Gödrei - LOG: print step.yml path FIX: removed global step info handling from local steps (2015 Dec 22)
* [64fd086] Viktor Benei - Merge pull request #167 from godrei/godep_update (2015 Dec 17)
* [136eb00] Krisztián Gödrei - godep update (2015 Dec 17)
* [0104b69] Viktor Benei - Merge pull request #166 from godrei/println_fix (2015 Dec 17)
* [4658ac9] Krisztián Gödrei - FIX: typo (2015 Dec 17)
* [75f6f52] Krisztián Gödrei - merge (2015 Dec 17)
* [e3c43be] Viktor Benei - Merge pull request #163 from godrei/step_info (2015 Dec 17)
* [d5cd881] Viktor Benei - Merge pull request #164 from godrei/typo (2015 Dec 17)
* [653b2d4] Krisztián Gödrei - FIX: typo Faild (2015 Dec 17)
* [34b647d] Krisztián Gödrei - printf fix (2015 Dec 17)
* [aa88983] Krisztián Gödrei - global step info handling (2015 Dec 16)
* [ff863dd] Viktor Benei - Merge pull request #162 from viktorbenei/master (2015 Dec 12)
* [ad697ee] Viktor Benei - godeps-update (2015 Dec 12)
* [9a31b91] Viktor Benei - Merge pull request #161 from godrei/share_fixes (2015 Dec 12)
* [6e82130] Krisztián Gödrei - step sharing improvements (2015 Dec 08)
* [5f6705b] Viktor Benei - Merge pull request #160 from viktorbenei/master (2015 Oct 05)
* [266202b] Viktor Benei - changelog fix (2015 Oct 02)
* [fb6b688] Viktor Benei - Merge pull request #159 from viktorbenei/master (2015 Oct 02)
* [91df01c] Viktor Benei - changelog format fix (2015 Oct 02)


## 0.9.17 (2015 Oct 02)

### Release Notes

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
* New command: `stepman collections` prints all the registered Step Lib collections.
* `stepman step-info` output now contains the input `default_value`, `value_options` and `is_expand` values and a couple more useful infos, like `source_code_url` and `support_url` of the Step.
* `stepman step-info` got a new option `--step-yml` flag, which allows printing step info from the specified `step.yml` directly (useful for local Step development).
* log improvements

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.17/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.16 -> 0.9.17

* [9377c16] Viktor Benei - Merge pull request #158 from viktorbenei/master (2015 Oct 02)
* [7f7d35a] Viktor Benei - v0.9.17 with changelog (2015 Oct 02)
* [67eaef1] Viktor Benei - godeps-update (2015 Oct 02)
* [df14870] Viktor Benei - Merge pull request #157 from gkiki90/changelog (2015 Oct 01)
* [ce4263c] Krisztian Goedrei - changelog (2015 Oct 01)
* [1f8bc5d] Viktor Benei - Merge pull request #156 from gkiki90/step_info (2015 Oct 01)
* [31e30a2] Krisztian Goedrei - step info (2015 Oct 01)
* [e6d6a43] Krisztian Goedrei - step title (2015 Oct 01)
* [2abd7bc] Viktor Benei - Merge pull request #155 from gkiki90/constructor (2015 Sep 30)
* [a9373b2] Krisztian Goedrei - json constructor (2015 Sep 30)
* [a715ec7] Viktor Benei - Merge pull request #153 from gkiki90/tool_mode (2015 Sep 29)
* [78f5f8a] Viktor Benei - Merge pull request #154 from gkiki90/log_fix (2015 Sep 29)
* [5e5e3bd] Krisztian Goedrei - log fix (2015 Sep 28)
* [f0004b4] Krisztian Goedrei - share tool mode (2015 Sep 28)
* [75ed40a] Viktor Benei - Merge pull request #152 from gkiki90/is_template (2015 Sep 25)
* [8e1eb34] Krisztian Goedrei - godep (2015 Sep 25)
* [aa55602] Krisztian Goedrei - require in tests (2015 Sep 25)
* [142c4b0] Viktor Benei - Merge pull request #151 from gkiki90/collections (2015 Sep 24)
* [481e495] Krisztian Goedrei - collections cmd (2015 Sep 24)
* [c1954a6] Viktor Benei - Merge pull request #150 from gkiki90/deps (2015 Sep 24)
* [267ec47] Krisztian Goedrei - refactor (2015 Sep 24)
* [a9c459c] Krisztian Goedrei - check only deps (2015 Sep 23)
* [6af6575] Viktor Benei - Merge pull request #149 from gkiki90/deps (2015 Sep 23)
* [e367567] Krisztian Goedrei - new deps model (2015 Sep 23)
* [90d4780] Krisztian Goedrei - new dep models (2015 Sep 23)
* [fec2552] Viktor Benei - Merge pull request #148 from gkiki90/step-info (2015 Sep 22)
* [d480239] Krisztian Goedrei - local step info (2015 Sep 22)
* [87fa652] Krisztian Goedrei - ci fix (2015 Sep 22)
* [11a545d] Krisztian Goedrei - godeps update (2015 Sep 22)
* [1f0b839] Krisztian Goedrei - ci fix, step info fix, godep (2015 Sep 22)
* [2fae766] Krisztian Goedrei - is expand (2015 Sep 22)
* [7a624d7] Viktor Benei - Merge pull request #147 from viktorbenei/master (2015 Sep 21)
* [fe2e5aa] Viktor Benei - Docker file : bitrise CLI version update (2015 Sep 21)
* [e60b1be] Viktor Benei - Merge pull request #146 from gkiki90/step_info (2015 Sep 21)
* [06115ee] Krisztian Goedrei - fix (2015 Sep 21)
* [aeb8cb0] Krisztian Goedrei - step info fix (2015 Sep 21)


## 0.9.16 (2015 Sep 17)

### Release Notes

* __BREAKING__ : `stepman step-info` command default output format changed from json to raw, for json output call `stepman step-info --format json`.
* step-info, step-list command get optional `--format` flag, which defines the output format (options: raw/json, default: raw).
* StepModel got new field `AssetURLs`, which holds the URI of step assets.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.16/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.15 -> 0.9.16

* [2515a68] Viktor Benei - Merge pull request #145 from gkiki90/changelog (2015 Sep 17)
* [64f73a2] Krisztian Goedrei - change log (2015 Sep 17)
* [b9a6c3e] Viktor Benei - Merge pull request #144 from gkiki90/model_fixes (2015 Sep 17)
* [6292594] Krisztian Goedrei - fix (2015 Sep 17)
* [8825b8e] Viktor Benei - Merge pull request #143 from gkiki90/asset_handling_fix (2015 Sep 17)
* [3057af5] Krisztian Goedrei - godep save (2015 Sep 17)
* [dfba88d] Krisztian Goedrei - merge (2015 Sep 17)
* [5b30563] Krisztian Goedrei - godeps (2015 Sep 17)
* [dc9abdf] Viktor Benei - Merge pull request #142 from gkiki90/format (2015 Sep 16)
* [63dfe83] Krisztian Goedrei - code cleaning (2015 Sep 16)
* [88c5e47] Krisztian Goedrei - models moved to models (2015 Sep 16)
* [0b6d099] Krisztian Goedrei - fix (2015 Sep 16)
* [03a8b5a] Krisztian Goedrei - step list, step info format (2015 Sep 16)
* [f5f8c07] Viktor Benei - Merge pull request #141 from viktorbenei/master (2015 Sep 16)
* [28b3fa6] Krisztian Goedrei - format fixes (2015 Sep 16)
* [a5cbdfd] Viktor Benei - start of v0.9.16 (2015 Sep 16)


## 0.9.15 (2015 Sep 16)

### Release Notes

* __New command__ : `stepman step-list` can be used to get a full list of available steps, in a specified library
* New Step property: `asset_urls`, auto-generated into the `spec.json` of the collection if `assets_download_base_uri` is defined in the collection's `steplib.yml`. This can be used to include assets URLs attached to a step, for example icons, primarily for UI tools/websites processing the `spec.json`.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.15/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.14 -> 0.9.15

* [514142d] Viktor Benei - Merge pull request #140 from viktorbenei/master (2015 Sep 16)
* [dc3bca4] Viktor Benei - v0.9.15, with a new changlog template (fixed `curl` call, to fail in case of network error) (2015 Sep 16)
* [6828714] Viktor Benei - script revisions: install bitrise script migrated into Dockerfile, create release with docker migrated into `bitrise.yml` (2015 Sep 16)
* [67de080] Viktor Benei - Merge pull request #139 from gkiki90/assets (2015 Sep 16)
* [4130c9d] Krisztian Goedrei - refactor (2015 Sep 16)
* [53ee421] Krisztian Goedrei - typo (2015 Sep 16)
* [47275e2] Viktor Benei - Merge pull request #138 from viktorbenei/master (2015 Sep 16)
* [98924ee] Viktor Benei - godeps-update (2015 Sep 16)
* [c7599d7] Krisztian Goedrei - assets fix (2015 Sep 15)
* [fa9f430] Krisztian Goedrei - assets (2015 Sep 15)
* [db19bd6] Viktor Benei - Merge pull request #136 from gkiki90/ci (2015 Sep 15)
* [48bacde] Krisztian Goedrei - ci fix (2015 Sep 15)
* [e1b6d7f] Krisztian Goedrei - slack fix (2015 Sep 15)
* [6bea395] Viktor Benei - Merge pull request #137 from gkiki90/step_list (2015 Sep 15)
* [f68c240] Krisztian Goedrei - step list (2015 Sep 14)
* [7abdce4] Krisztian Goedrei - fix (2015 Sep 14)
* [50e7cd6] Krisztian Goedrei - new ci (2015 Sep 14)
* [1db78f4] Viktor Benei - Merge pull request #135 from viktorbenei/master (2015 Sep 08)
* [0307a89] Viktor Benei - start of v0.9.15 (2015 Sep 08)


## 0.9.14 (2015 Sep 08)

### Release Notes

* Internal revisions and a directory copy fix

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.14/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.13 -> 0.9.14

* [96668a7] Viktor Benei - Merge pull request #134 from viktorbenei/master (2015 Sep 08)
* [330f91c] Viktor Benei - v0.9.14 (2015 Sep 08)
* [b7b83da] Viktor Benei - Merge pull request #133 from viktorbenei/master (2015 Sep 08)
* [088c026] Viktor Benei - full godeps-update (2015 Sep 08)
* [5a8c5f8] Viktor Benei - Merge pull request #132 from viktorbenei/master (2015 Sep 08)
* [c44283d] Viktor Benei - godeps-update : for CopyDir #fix (2015 Sep 08)
* [37dea3c] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Sep 08)
* [00a359e] Viktor Benei - a bit more model property order change, for serialization (2015 Sep 08)
* [e4afd1a] Viktor Benei - StepModel: order change, summary is now before description & a bit of commenting (2015 Sep 08)
* [e49f70b] Viktor Benei - Merge pull request #131 from viktorbenei/master (2015 Sep 07)
* [4baef2a] Viktor Benei - start of v0.9.14 (2015 Sep 07)


## 0.9.13 (2015 Sep 07)

### Release Notes

* __NEW__ command : `stepman step-info` which prints a JSON information about the specified step - useful for tools.
* `stepman audit` __fix__ : better validation.
* internal revisions, mainly for better data validation.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.13/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.12 -> 0.9.13

* [4d134ad] Viktor Benei - Merge pull request #130 from viktorbenei/master (2015 Sep 07)
* [1823a31] Viktor Benei - Dockerfile update for Linux builds (2015 Sep 07)
* [1304f35] Viktor Benei - v0.9.13 changelog (2015 Sep 07)
* [e540614] Viktor Benei - godeps-update : include sub packages (2015 Sep 07)
* [e722bdb] Viktor Benei - Merge pull request #129 from gkiki90/step_info (2015 Sep 03)
* [be9c102] Krisztian Goedrei - step_info (2015 Sep 03)
* [3f68e2a] Viktor Benei - Merge pull request #128 from gkiki90/default_steplib_fix (2015 Sep 03)
* [66b9f2c] Viktor Benei - Merge pull request #127 from gkiki90/audit_fix (2015 Sep 03)
* [fc89b93] Krisztian Goedrei - bitrise yml fix (2015 Sep 02)
* [89a896c] Krisztian Goedrei - steplib fix (2015 Sep 02)
* [fee931b] Krisztian Goedrei - audit test (2015 Sep 02)
* [978f419] Krisztian Goedrei - audit checks required fields (2015 Sep 02)
* [f25b954] Viktor Benei - Merge pull request #125 from gkiki90/model_fix (2015 Sep 02)
* [e5ca6ec] Krisztian Goedrei - test fix (2015 Sep 02)
* [44ee56b] Krisztian Goedrei - PR fix (2015 Sep 02)
* [157719e] Krisztian Goedrei - step env validate fix (2015 Sep 02)
* [1c0c811] Viktor Benei - Merge pull request #126 from viktorbenei/master (2015 Sep 02)
* [8068862] Viktor Benei - `golint` now fails CI if finds any issue (2015 Sep 02)
* [dbf5086] Krisztian Goedrei - model fix (2015 Sep 02)
* [585f214] Viktor Benei - Merge pull request #124 from viktorbenei/master (2015 Aug 31)
* [6a0edcd] Viktor Benei - updated bitrise-CLI install (2015 Aug 31)
* [0d87f27] Viktor Benei - Merge pull request #123 from viktorbenei/master (2015 Aug 31)
* [dae9a09] Viktor Benei - start of v0.9.13 (2015 Aug 31)


## 0.9.12 (2015 Aug 31)

### Release Notes

* __NEW__ Step property: `published_at`
* Log format revision, unified with `envman` and `bitrise` CLI
* __FIX__ Version compare fixed in `go-utils`

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.12/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.11 -> 0.9.12

* [cbb9e52] Viktor Benei - Merge pull request #122 from viktorbenei/master (2015 Aug 31)
* [11e9e4a] Viktor Benei - changelog : version compare fix note (2015 Aug 31)
* [65eeb5e] Viktor Benei - godeps-update : version compare #fix (2015 Aug 31)
* [1a30e68] Viktor Benei - changelog v0.9.12 (2015 Aug 31)
* [e1a0c95] Viktor Benei - Merge pull request #121 from gkiki90/published_at (2015 Aug 31)
* [3897800] Krisztian Goedrei - godeps-update (2015 Aug 31)
* [bd3476e] Krisztian Goedrei - published at (2015 Aug 31)
* [3568869] Viktor Benei - Merge pull request #120 from gkiki90/published_at_fix (2015 Aug 31)
* [7400b7d] Krisztian Goedrei - change published_at type to time (2015 Aug 31)
* [825ffa6] Viktor Benei - Merge pull request #119 from gkiki90/1_0_0_models (2015 Aug 31)
* [8cc3391] Krisztian Goedrei - new step field published_at (2015 Aug 31)
* [2a680d5] Viktor Benei - Merge pull request #118 from viktorbenei/master (2015 Aug 28)
* [31ef868] Viktor Benei - godeps-update (2015 Aug 28)
* [c6eaf8f] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Aug 28)
* [b52d5d9] Viktor Benei - Merge pull request #117 from gkiki90/ci (2015 Aug 27)
* [54731cf] Krisztian Goedrei - ci (2015 Aug 27)
* [6198511] Viktor Benei - Merge pull request #116 from gkiki90/master (2015 Aug 27)
* [05b0ec2] Krisztian Goedrei - force color log (2015 Aug 27)
* [0d53aef] Viktor Benei - godeps-update (2015 Aug 26)
* [fb5af44] Viktor Benei - Merge pull request #115 from gkiki90/init_fix_&_path_fix (2015 Aug 26)
* [c5fdb26] Krisztian Goedrei - init fix & path.Join fixes (2015 Aug 25)
* [11a6bc2] Viktor Benei - Merge pull request #114 from viktorbenei/master (2015 Aug 24)
* [c76022b] Viktor Benei - start of v0.9.12 (2015 Aug 24)


## 0.9.11 (2015 Aug 24)

### Release Notes

* The `share` command was extended with an optional `audit` call, to test the integrity or the StepLib before finishing the Step share.
* Step inputs can now (optionally) contain a `summary` as well as a `title` and `description`.
* A couple of minor revision.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.11/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.10 -> 0.9.11

* [c11cd3e] Viktor Benei - Merge pull request #113 from viktorbenei/master (2015 Aug 24)
* [495a55d] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Aug 24)
* [246c074] Viktor Benei - changelog v0.9.11 (2015 Aug 24)
* [957ae5a] Viktor Benei - Merge pull request #112 from viktorbenei/master (2015 Aug 24)
* [bb4e421] Viktor Benei - godeps-update (2015 Aug 24)
* [2718c3e] Viktor Benei - Merge pull request #111 from gkiki90/commit_hash_test (2015 Aug 19)
* [1cd1e18] Krisztian Goedrei - commit hash test (2015 Aug 19)
* [1143aa2] Viktor Benei - Merge pull request #110 from gkiki90/share_audit (2015 Aug 18)
* [29d9f64] Krisztian Goedrei - share audit (2015 Aug 18)
* [5801b23] Viktor Benei - Merge pull request #109 from viktorbenei/master (2015 Aug 17)
* [3d14f88] Viktor Benei - godeps-update : envman opts parsing map[string]interface fix (2015 Aug 17)
* [7c37f6b] Viktor Benei - Update README.md (2015 Aug 14)
* [723c9f8] Viktor Benei - Update README.md (2015 Aug 14)
* [db8327b] Viktor Benei - Merge pull request #108 from viktorbenei/master (2015 Aug 14)
* [89591e2] Viktor Benei - switch to bitrise 0.9.10 (2015 Aug 14)
* [a1d8d76] Viktor Benei - Merge pull request #107 from viktorbenei/master (2015 Aug 14)
* [a960a88] Viktor Benei - start of v0.9.11 (2015 Aug 14)


## 0.9.10 (2015 Aug 14)

### Release Notes

* __FIX__ : if `activate` called with the `--update` flag and the step is not found in the local StepLib collection it'll do an `update` and re-check the step. Now works for both if you specify a version for the step or if you don't (if you use the "latest" version).
* Improved guide for `share`

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.10/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.9 -> 0.9.10

* [deee980] Viktor Benei - Merge pull request #106 from viktorbenei/master (2015 Aug 14)
* [18ee4b3] Viktor Benei - changelog 0.9.10 (2015 Aug 14)
* [d2e17e3] Viktor Benei - Merge pull request #105 from viktorbenei/master (2015 Aug 14)
* [f0f798d] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Aug 14)
* [f0cfdd7] Viktor Benei - godep update (2015 Aug 14)
* [6143573] Viktor Benei - Merge pull request #103 from gkiki90/StepLib_auto_update_fix (2015 Aug 14)
* [5570b7f] Krisztian Goedrei - auto update fix (2015 Aug 14)
* [56a6954] Viktor Benei - Merge pull request #104 from viktorbenei/master (2015 Aug 14)
* [7f76c3b] Viktor Benei - the _step_template folder moved into the main `bitrise` repo (2015 Aug 14)
* [e33f55e] Viktor Benei - Merge pull request #102 from viktorbenei/master (2015 Aug 14)
* [aeadc78] Viktor Benei - prepare for `bitrise setup --minimal` (2015 Aug 14)
* [31cbcd7] Viktor Benei - Merge pull request #100 from gkiki90/share_finish_fix (2015 Aug 14)
* [34e23d3] Viktor Benei - Merge pull request #101 from viktorbenei/master (2015 Aug 14)
* [c94962e] Viktor Benei - install bitrise osx script (2015 Aug 14)
* [ffec384] Krisztian Goedrei - share finish msg (2015 Aug 14)
* [f583ac2] Viktor Benei - Merge pull request #99 from gkiki90/share_fixes (2015 Aug 14)
* [760e0e2] Krisztian Goedrei - commit msg (2015 Aug 14)
* [f8d8a13] Viktor Benei - Merge pull request #98 from gkiki90/step_template (2015 Aug 13)
* [79e76ed] Krisztian Goedrei - step template (2015 Aug 13)
* [dd50696] Viktor Benei - Merge pull request #97 from viktorbenei/master (2015 Aug 13)
* [815f792] Viktor Benei - start of v0.9.10 (2015 Aug 13)


## 0.9.9 (2015 Aug 13)

### Release Notes

* __NEW__ : `stepman audit` command, to help you solve issues with your Step before sharing it with others.
* Lots of code revision.

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.9/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.8 -> 0.9.9

* [9f18de6] Viktor Benei - Merge pull request #96 from gkiki90/audit (2015 Aug 13)
* [e65a9ab] Krisztian Goedrei - audit (2015 Aug 13)
* [2eb6c7e] Viktor Benei - Merge pull request #95 from gkiki90/go-utils_cli-fix (2015 Aug 13)
* [8ee8365] Krisztian Goedrei -  cli fixes, go-util, godep-update (2015 Aug 13)
* [688412c] Viktor Benei - Merge pull request #94 from viktorbenei/master (2015 Aug 13)
* [05b7a46] Viktor Benei - setup: fail if step.json can't be generated (2015 Aug 13)
* [a80fe27] Viktor Benei - Merge pull request #93 from viktorbenei/master (2015 Aug 12)
* [7a88381] Viktor Benei - start of v0.9.9 (2015 Aug 12)


## 0.9.8 (2015 Aug 12)

### Release Notes

* __BREAKING__ : `step.yml` shared in Step Libraries / Step Collections now have to include a `commit` (hash) property inside the `source` property, for better version validation (version tag have to match this commit hash)!
    * You should switch to the new, final default StepLib, hosted on GitHub, which contains these commit hashes and works with stepman 0.9.8! URL: https://github.com/bitrise-io/bitrise-steplib
    * We'll soon (in about 1 day) start to accept Step contributions to this new StepLib!
* __NEW__ : built in commands to `share` a new step into a StepLib!
* Option to `setup` a local StepLib (use a local path as source instead of a remote git url)
* Delete command : removes the specified collection from the local cache completely.
* Lots of code revision & minor fixes

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.8/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.7 -> 0.9.8

* [83bb2fd] Viktor Benei - Merge pull request #92 from viktorbenei/master (2015 Aug 12)
* [9f48e39] Viktor Benei - godeps-update (2015 Aug 12)
* [527a837] Viktor Benei - Merge pull request #91 from viktorbenei/master (2015 Aug 12)
* [8987bb3] Viktor Benei - finishing touches on stepman share guide texts (2015 Aug 12)
* [caeeece] Viktor Benei - minor stepman share create guide text formatting revision (2015 Aug 12)
* [ff504df] Viktor Benei - Merge pull request #90 from viktorbenei/master (2015 Aug 12)
* [f13b32d] Viktor Benei - link fixes in share guide (2015 Aug 12)
* [7b28453] Viktor Benei - Merge pull request #89 from gkiki90/share_fix (2015 Aug 12)
* [6fa4aae] Krisztian Goedrei - godep-update (+3 squashed commits) Squashed commits: [2d1f855] share fixes [e189be4] share fixes [a7491c0] code style, fixes (2015 Aug 12)
* [94f0799] Viktor Benei - Merge pull request #88 from viktorbenei/master (2015 Aug 12)
* [1d45184] Viktor Benei - minimal logging change (2015 Aug 12)
* [2ccf68c] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Aug 12)
* [a6f50d0] Viktor Benei - successfully downloaded step.zip : log now contains the ZIP's URL (2015 Aug 12)
* [60195a3] Viktor Benei - Merge pull request #87 from gkiki90/master (2015 Aug 11)
* [14fe2a0] Krisztian Goedrei - downloadAndUnzip quick fix (2015 Aug 11)
* [be45cd9] Viktor Benei - Merge pull request #86 from viktorbenei/master (2015 Aug 11)
* [984fe37] Viktor Benei - error msg typo fix (2015 Aug 11)
* [bc65fae] Viktor Benei - Merge pull request #85 from gkiki90/validate_step_source (2015 Aug 11)
* [ddc8a59] Krisztian Goedrei - PR fixes (+1 squashed commit) Squashed commits: [19f5702] godep-update (+2 squashed commits) Squashed commits: [f8a631b] code style [4726671] validate step source (2015 Aug 11)
* [4041b77] Viktor Benei - Merge pull request #84 from viktorbenei/feature/fix-source-commit-is-saved-with-newlines (2015 Aug 11)
* [235d9dc] Viktor Benei - fixed the git commit newline fix -> was fixed in go-utils package; we godep-updated to it (2015 Aug 11)
* [008fe7e] Viktor Benei - Merge pull request #83 from viktorbenei/master (2015 Aug 11)
* [b0c3945] Viktor Benei - go-utils migrations (2015 Aug 11)
* [53fda3d] Viktor Benei - Merge pull request #82 from viktorbenei/master (2015 Aug 11)
* [9c65242] Viktor Benei - go-pathutil now used from the new go-utils repo (2015 Aug 11)
* [5e20040] Viktor Benei - Merge pull request #81 from gkiki90/master (2015 Aug 10)
* [47f0979] Krisztian Goedrei - godep update (2015 Aug 10)
* [7a51722] Viktor Benei - Merge pull request #80 from gkiki90/model_fix (2015 Aug 10)
* [6f96425] Krisztian Goedrei - go-util for prts (+5 squashed commits) Squashed commits: [800cd23] godep-update [8243f14] ptr with go-utils [ae74f8d] godep-update [fc9e007] test fixes, import fixes [23d2835] fixed fill missing defaults (2015 Aug 10)
* [5a2bdc5] Viktor Benei - Merge pull request #78 from viktorbenei/master (2015 Aug 08)
* [5049b0b] Viktor Benei - collection delete - route save fix (2015 Aug 08)
* [e226718] Viktor Benei - Merge pull request #75 from gkiki90/master (2015 Aug 08)
* [37234b7] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/stepman (2015 Aug 08)
* [26deca3] Viktor Benei - Merge pull request #74 from viktorbenei/master (2015 Aug 08)
* [c3dd038] Viktor Benei - source is not required anymore (+1 squashed commit) Squashed commits: [45ff51a] simple step_template (2015 Aug 08)
* [c4eefe8] Viktor Benei - Merge pull request #71 from gkiki90/share (2015 Aug 08)
* [c8cddcf] Krisztian Goedrei - final (!!) fixes (2015 Aug 08)
* [f6d505c] Krisztian Goedrei - code style (2015 Aug 08)
* [52be172] Krisztian Goedrei - step validation fix (2015 Aug 08)
* [a2d6dc9] Krisztian Goedrei - tmp dir fix (2015 Aug 08)
* [5a1ea95] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/stepman into share (2015 Aug 08)
* [fec2118] Krisztian Goedrei - ReGenerateStepSpec (2015 Aug 08)
* [cba3d6e] Viktor Benei - Merge pull request #73 from viktorbenei/master (2015 Aug 08)
* [0ad001e] Viktor Benei - don't fail in delete if doesn't exist (2015 Aug 08)
* [50cdb61] Viktor Benei - log fix in activate (2015 Aug 08)
* [ee46b83] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/stepman (2015 Aug 08)
* [6d8cd16] Krisztian Goedrei - log fixes (2015 Aug 08)
* [fb34c39] Viktor Benei - Merge pull request #72 from viktorbenei/master (2015 Aug 08)
* [1bb6c4d] Krisztian Goedrei - PR fixes (2015 Aug 08)
* [5df7155] Viktor Benei - Godeps update (2015 Aug 08)
* [df00ed6] Krisztian Goedrei - missing errcheck (2015 Aug 08)
* [4ed716f] Krisztian Goedrei - flag message revision (2015 Aug 08)
* [307ed02] Krisztian Goedrei - godep update (2015 Aug 08)
* [594ea48] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/stepman into share (2015 Aug 08)
* [fefb547] Krisztian Goedrei - Merge branch 'source_commithash' into share (2015 Aug 08)
* [fc17a56] Viktor Benei - Merge pull request #70 from viktorbenei/delete-action (2015 Aug 08)
* [3d72135] Viktor Benei - delete command, to delete a collection (+1 squashed commit) Squashed commits: [386cdb3] updated default-steplib-source in bitrise.yml (2015 Aug 08)
* [dc35836] Viktor Benei - Merge pull request #69 from viktorbenei/master (2015 Aug 08)
* [76e4f77] Viktor Benei - couple of logging fixes and GitCheckout hash handling fix + a bit more verbose log if git clone fails (2015 Aug 08)
* [dad904c] Viktor Benei - Merge pull request #68 from gkiki90/source_commithash (2015 Aug 08)
* [5c6d17f] Krisztian Goedrei - source increased with commit hash (2015 Aug 08)
* [f8da5e2] Krisztian Goedrei - share finish in progress (2015 Aug 08)
* [ba4dfa4] Krisztian Goedrei - share finish in progress (2015 Aug 07)
* [d12f0f6] Krisztian Goedrei - merge into source_commit (2015 Aug 07)
* [8fcfb43] Krisztian Goedrei - create in progress (2015 Aug 07)
* [dd601a8] Krisztian Goedrei - source commit fixes (2015 Aug 07)
* [400eb16] Krisztian Goedrei - merge with source_commit branch (2015 Aug 07)
* [cb7a06b] Krisztian Goedrei - validate commithash (2015 Aug 07)
* [0f39a6b] Krisztian Goedrei - source increased with commit hash (2015 Aug 07)
* [6f4e25a] Krisztian Goedrei - commands (2015 Aug 07)
* [9471caa] Krisztian Goedrei - share, share start cmd, remove route fix (2015 Aug 07)
* [b07b7c0] Viktor Benei - Merge pull request #67 from viktorbenei/master (2015 Aug 07)
* [18f23a7] Viktor Benei - new options in 'setup' action : use a local path as source of the collection + optionally specify a spec.json path, to copy the successfully generated spec.json to (2015 Aug 06)
* [4701020] Viktor Benei - Merge pull request #66 from viktorbenei/master (2015 Aug 06)
* [f0a5c41] Viktor Benei - removed collection ENV manual handling, it's done through codegangsta/cli automatically (2015 Aug 06)
* [4228356] Viktor Benei - Merge pull request #65 from viktorbenei/master (2015 Aug 05)
* [0812440] Viktor Benei - start of v0.9.8 (2015 Aug 05)


## 0.9.7 (2015 Aug 05)

### Release Notes

* Minor fixes

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.7/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.6 -> 0.9.7

* [319ee3c] Viktor Benei - Merge pull request #64 from gkiki90/master (2015 Aug 05)
* [005a4be] Krisztian Goedrei - changelog (2015 Aug 05)
* [2ecd4b8] Viktor Benei - Merge pull request #63 from gkiki90/master (2015 Aug 05)
* [76b9edc] Krisztian Goedrei - Merge branch 'master' of github.com:bitrise-io/stepman (2015 Aug 05)
* [4e8da36] Krisztian Goedrei - flag fixes (2015 Aug 05)
* [aff016d] Viktor Benei - Merge pull request #62 from viktorbenei/master (2015 Aug 05)
* [3de6aeb] Viktor Benei - changelog template fix (2015 Aug 04)
* [2d8c783] Viktor Benei - Merge pull request #60 from viktorbenei/master (2015 Aug 03)
* [079dcb3] Viktor Benei - Merge pull request #61 from gkiki90/master (2015 Aug 03)
* [9e8df0d] Krisztian Goedrei - godep-update (2015 Aug 03)
* [1d95dbb] Viktor Benei - start of v0.9.7 (2015 Aug 03)


## 0.9.6 (2015 Aug 03)

### Release Notes

* Environment models moved to Envman.
* Less verbose log at first setup of Steplib.
* Dependencies added to StepModel (currently supported dependency manager: brew)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.6/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.5 -> 0.9.6

* [856b2ca] Viktor Benei - Merge pull request #59 from gkiki90/release (2015 Aug 03)
* [637ff38] Krisztian Goedrei - auto changelog (2015 Aug 03)
* [033a3c6] Viktor Benei - Merge pull request #58 from gkiki90/master (2015 Aug 03)
* [2096acf] Krisztian Goedrei - models quick fix (2015 Aug 03)
* [b3d1905] Viktor Benei - Merge pull request #57 from gkiki90/master (2015 Aug 03)
* [5124b24] Krisztian Goedrei - quick fix typo (2015 Aug 03)
* [9ed8390] Viktor Benei - Merge pull request #56 from gkiki90/step_requirements (2015 Aug 03)
* [dd751a6] Krisztian Goedrei - models fixes (2015 Aug 03)
* [387241b] Krisztian Goedrei - ci fix (2015 Aug 03)
* [c310bb8] Krisztian Goedrei - _script fixes (2015 Aug 03)
* [81744fc] Krisztian Goedrei - log fixes (2015 Aug 03)
* [c87397b] Krisztian Goedrei - typo fix (2015 Jul 31)
* [542cbd9] Krisztian Goedrei - refactor (2015 Jul 31)
* [ed1fcf8] Krisztian Goedrei - godep update (2015 Jul 30)
* [3ce4c0e] Krisztian Goedrei - log fix (2015 Jul 30)
* [a98e2c1] Krisztian Goedrei - stepman test (2015 Jul 30)
* [e4387df] Krisztian Goedrei - code style, models_methods_test (2015 Jul 30)
* [271e5c1] Krisztian Goedrei - typo (2015 Jul 30)
* [27a3361] Viktor Benei - Merge pull request #54 from gkiki90/new_envman_models (2015 Jul 29)
* [6667fad] Krisztian Goedrei - refactor (2015 Jul 29)
* [5cc73ab] Krisztian Goedrei - godep update (2015 Jul 29)
* [9ffb824] Krisztian Goedrei - code cleaning (2015 Jul 29)
* [a69696b] Krisztian Goedrei - new envman modesl (2015 Jul 29)
* [23abaa4] Viktor Benei - Merge pull request #53 from viktorbenei/master (2015 Jul 28)
* [860913c] Viktor Benei - start of v0.9.6 (2015 Jul 28)


## 0.9.5 (2015 Jul 28)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.5/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.4 -> 0.9.5

* [d66cfb3] Viktor Benei - Merge pull request #52 from gkiki90/master (2015 Jul 28)
* [8d9bc23] Krisztian Goedrei - env validate fix (2015 Jul 28)
* [0a3d6c2] Viktor Benei - Merge pull request #51 from gkiki90/normalize_env (2015 Jul 28)
* [e8c6cbd] Krisztian Goedrei - normalize env fix (2015 Jul 28)
* [658ae39] Krisztian Goedrei - normalize envs (2015 Jul 28)
* [55e3a3d] Viktor Benei - Merge pull request #50 from viktorbenei/master (2015 Jul 28)
* [8ccbb60] Viktor Benei - bit of more verbose error msg for Validate (2015 Jul 28)
* [2abf980] Viktor Benei - Merge pull request #49 from viktorbenei/master (2015 Jul 24)
* [c11b5e9] Viktor Benei - FillMissingDeafults once again fills every empty missing property of Step and EnvItem (2015 Jul 24)
* [f9af4f7] Viktor Benei - Merge pull request #48 from viktorbenei/run_if (2015 Jul 24)
* [7393e5a] Viktor Benei - removed unnecessary, optional step property fill-missing-defaults (2015 Jul 24)
* [34f7b25] Viktor Benei - added RunIf property to Step (2015 Jul 24)
* [d9ac8a3] Viktor Benei - Merge pull request #47 from viktorbenei/master (2015 Jul 24)
* [bc10710] Viktor Benei - Install instructions now points to /releases (2015 Jul 24)
* [be47eb8] Viktor Benei - start of v0.9.5 (2015 Jul 24)
* [7fa41b4] Viktor Benei - Merge pull request #46 from viktorbenei/master (2015 Jul 24)
* [bdef4d5] Viktor Benei - announce-release workflow fix (2015 Jul 24)
* [c9fb4cb] Viktor Benei - bitrise.yml fix (2015 Jul 24)
* [0ff86f4] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Jul 24)
* [515a819] Viktor Benei - bitrise.yml update (2015 Jul 24)


## 0.9.4 (2015 Jul 24)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.4/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.3 -> 0.9.4

* [c161122] Viktor Benei - Merge pull request #45 from viktorbenei/master (2015 Jul 24)
* [3400bf2] Viktor Benei - just a minor Default.. usage fix (2015 Jul 24)
* [0e20efb] Viktor Benei - IsNotImportant renamed to IsSkippable (2015 Jul 24)
* [f5a9485] Viktor Benei - Merge pull request #43 from gkiki90/routing_fix (2015 Jul 24)
* [1bbefb5] Krisztian Goedrei - PR fixes (2015 Jul 24)
* [4efc23b] Viktor Benei - Merge pull request #44 from viktorbenei/master (2015 Jul 23)
* [3be3376] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Jul 23)
* [033d5fc] Viktor Benei - start of v0.9.4 (2015 Jul 23)
* [1743c38] Viktor Benei - Install - 0.9.3 (2015 Jul 23)
* [f5902fc] Krisztian Goedrei - routing fixes (2015 Jul 23)
* [3a3c6a1] Krisztian Goedrei - fixed setup (2015 Jul 23)


## 0.9.3 (2015 Jul 23)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.3/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.2 -> 0.9.3

* [e28b2d1] Viktor Benei - Merge pull request #42 from viktorbenei/better_version_compare (2015 Jul 23)
* [df42234] Viktor Benei - added missing return err (2015 Jul 23)
* [9324630] Viktor Benei - better "CompareVersions" method (2015 Jul 23)
* [674a69b] Viktor Benei - Merge pull request #41 from viktorbenei/master (2015 Jul 22)
* [fff943b] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Jul 22)
* [301abc7] Viktor Benei - start of v0.9.3 (2015 Jul 22)
* [38fd4e4] Viktor Benei - Install - 0.9.2 (2015 Jul 22)


## 0.9.2 (2015 Jul 22)

### Install or upgrade

To install this version, run the following commands (in a bash shell):

```
curl -fL https://github.com/bitrise-io/stepman/releases/download/0.9.2/stepman-$(uname -s)-$(uname -m) > /usr/local/bin/stepman
```

Then:

```
chmod +x /usr/local/bin/stepman
```

That's all, you're ready to call `stepman`!

### Release Commits - 0.9.1 -> 0.9.2

* [08cfef7] Viktor Benei - Merge pull request #40 from viktorbenei/master (2015 Jul 22)
* [b819005] Viktor Benei - bitrise.yml : send a Slack msg when release is ready (2015 Jul 22)
* [5040a2b] Viktor Benei - Merge pull request #36 from gkiki90/help_messages (2015 Jul 22)
* [8c4a429] Krisztian Goedrei - message fixes (2015 Jul 22)
* [b741ecf] Viktor Benei - Merge pull request #38 from gkiki90/update_if_needed (2015 Jul 22)
* [bed8ad2] Krisztian Goedrei - code cleaning (2015 Jul 22)
* [5be0c4c] Viktor Benei - Merge pull request #39 from viktorbenei/master (2015 Jul 22)
* [6f9682f] Viktor Benei - minimal bitrise.yml change, for format version upgrade & a minimal revision of path gen (2015 Jul 22)
* [6d02511] Viktor Benei - Install command syntax change (2015 Jul 22)
* [65a8c8c] Viktor Benei - Merge pull request #37 from viktorbenei/master (2015 Jul 22)
* [d7191db] Viktor Benei - Merge branch 'master' of https://github.com/bitrise-io/stepman (2015 Jul 22)
* [ce7a4cc] Viktor Benei - start of v0.9.2 - bitrise.yml now contains a 'create-release' workflow (2015 Jul 22)
* [3dffba2] Krisztian Goedrei - update flag (2015 Jul 22)
* [2bc4fe6] Viktor Benei - base bitrise.yml for create-release (2015 Jul 22)
* [03b2c73] Viktor Benei - Install notes for v0.9.1 (2015 Jul 22)
* [4946d92] Krisztian Goedrei - typo (2015 Jul 22)
* [9a8307a] Krisztian Goedrei - help messages (2015 Jul 22)


-----------------

Generated at: 2017 Aug 07
