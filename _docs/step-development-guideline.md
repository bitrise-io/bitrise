# Step Development Guideline

## Never depend on Environment Variables in your Step

You should expose every outside variable as an input of your step,
and just set the default value to the Environment Variable you want to use in the `step.yml`.

An example:

The Xcode Archive step generates a `$BITRISE_IPA_PATH` output environment variable.
**You should not** use this environment variable in your Step's code directly,
instead you should declare an input for your Step in `step.yml` and just set the default
value to `$BITRISE_IPA_PATH`. Example:

```
- ipa_path: "$BITRISE_IPA_PATH"
  opts:
      title: "IPA path"
```

After this, in your Step's code you can expect that the `$ipa_path` Environment Variable will
contain the value of the IPA path.

By declaring every option as an input you make it easier to test your Step,
and you also let the user of your Step to easily declare these inputs,
instead of searching in the code for the required Environment Variable.

### Secret environment variables in Steps 

You can mark Step inputs as **Sensitive** to make sure their values do not get exposed. Sensitive inputs only accept [Secrets](/bitrise-cli/secrets/) - secret environment variables - as values. This ensures they are not visible in build logs. 

To mark a Step input as sensitive, use the `is_sensitive` property. It has two values: `true` and `false`. 

Please note that if you mark an input as sensitive, the `is_expand` property of the input also must be `true`!

```yaml
inputs:
  - certificate_urls: $BITRISE_CERTIFICATE_URL
    opts:
      title: "Certificate URL"
      is_sensitive: true
```

## Do not use submodules, or require any other resource downloaded on-demand

As a Step runs frequently in a CI / automation environment you should try to make your Step as stable as possible.
This includes the resources / tools used by your Step as well, not just the core code.

If your Step depends on another tool, which have to be downloaded on-demand, during the execution
of your Step, there's a chance that even your Step was retrieved correctly but the
resource it tries to download just fails because of a network, authorization or other error.

You should try to include everything what's required for your Step into the Step's repository.
In case of submodules, you should rather include the content of the other repository,
instead of actually using it as a submodule.

The only exception is the dependencies you can fetch from an OS dependency manager,
on Debian systems you can use `apt-get` and on OS X you can use `brew`.
You can declare these dependencies in your `step.yml`, with the `deps` property,
and `bitrise` will manage to call the dependency manager to install the dependency,
and will fail before the Step execution in case it can't retrieve the dependency.


## Step id naming convention

Use hyphen (`-`) separated words for you step id, like: `set-ios-bundle-identifier`, `xcode-archive-mac`, ...


## Input naming convention

Use lower case [snake case](https://en.wikipedia.org/wiki/Snake_case) style input IDs, e.g. `input_path`.

### Inputs which can accept a list of values

You should postfix the input ID with `_list` (e.g. `input_path_list`), and expect the values to be provided as a pipe character separated list (e.g. `first value|second value`). This is not a hard requirement, but a strong suggestion. This means that you should prefer this solution unless you really need to use another character for separating values. Based on our experience the pipe character (`|`) works really well as a universal separator character, as it's quite rare in input values (compared to `,`, `;`, `=` or other more common separator characters).

**As a best practice you should filter out empty items**, so that `first value||second value` or even

```
first value|       |second value
```

is treated the same way as `first value|second value`. Again, not a hard requirement, but based on our experience this is the most reliable long term solution.


## Output naming convention

Use all-upper-case [snake case](https://en.wikipedia.org/wiki/Snake_case) style output IDs, e.g. `OUTPUT_PATH`.

### List of values in outputs

You should postfix the output ID with `_LIST` (e.g. `OUTPUT_PATH_LIST`), and provide the values as a pipe separated list (e.g. `first value|second value`). This is not a hard requirement, but a strong suggestion. This means that you should prefer this solution unless you really need to use another character for separating values. Based on our experience the pipe character (`|`) works really well as a universal separator character, as it's quite rare in output values (compared to `,`, `;`, `=` or other more common separator characters).


## Version naming convention

You should use [semantic versioning](http://semver.org/) (MAJOR.MINOR.PATCH) for your step. For example: `1.2.3`.


## Step Grouping convention

You can use `project_type_tags` and `type_tags` to group/categorise your steps.

`project_type_tags` are used to control if the step is available/useful for the given project type.

Available `project_type_tags`:

- ios
- macos
- android
- xamarin
- react-native
- cordova
- ionic
- flutter
- fastlane
- web

_If step is available for all project types, do not specify project_type_tags, otherwise specify every project types, with which the step can work._

`type_tags` are used to categorise the steps based on it's functionality.

Available `type_tags`:

- access-control
- artifact-info
- installer
- deploy
- utility
- dependency
- code-sign
- build
- test
- notification

_Every step should have at least one type_tag, if you feel you would need a new one, or update an existing's name, please [create a github issue](https://github.com/bitrise-io/bitrise/issues/new), with your suggestion._

## Step icon convention

The step's icon should match with the following criterias:

- NOT transparent background color
- size: 256 * 256 px
- margin: 60 px
- format: SVG

Submit your step's icon by:

- adding the svg file into your steplib fork repo at: STEPLIB_FORK_ROOT/steps/YOUR_STEP_ID/assets/icon.svg
- createing a new pull request to the [steplib repo](https://github.com/bitrise-io/bitrise-steplib)
