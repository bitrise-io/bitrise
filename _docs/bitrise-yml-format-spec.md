# bitrise.yml format specification / reference

## Minimal bitrise.yml

The bare minimum `bitrise.yml` is:

```
format_version: 2
```

Minimum `bitrise.yml` for a single (no-op) workflow:

```
format_version: 2
workflows:
  test:
```

## Top level bitrise.yml properties

- `format_version` : this property declares the minimum Bitrise CLI format version.
  You can get your Bitrise CLI's supported highest format version with: `bitrise version --full`.
  If you set the `format_version` to `2` that means that Bitrise CLI versions which
  don't support the format version `2` or higher won't be able to run the configuration.
  This is important if you use features which are not available in older Bitrise CLI versions.
- `default_step_lib_source` : specifies the source to use when no other source is defined for a step.
- `project_type` : defines your source project's type.
- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `app` : global, "app" specific configurations.
- `trigger_map` : Trigger Map definitions.
- `workflows` : workflow definitions.

## App properties

- `envs` : configuration global environment variables list
- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._

##  Trigger Map

Trigger Map is a list of Trigger Map Items. The elements of the list are processed ordered. If one item matches to the current git event, the item defined workflow will be run.

## Trigger Map Item

Trigger Map Item defines what kind of git event should trigger which workflow.

The Trigger Map Item layout: 

```
git_event_property: pattern
workflow: workflow_id
```

Available trigger events ( with properties ):

- Code Push (`push_branch`)
- Pull Request (`pull_request_source_branch`, `pull_request_target_branch`)
- Creating Tag (`tag`)

## Workflow properties

- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `before_run` : list of workflows to execute before this workflow
- `after_run` : list of workflows to execute after this workflow
- `envs` : workflow defined environment variables list
- `steps` : workflow defined step list

## Step properties

- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `website` : official website of the step / service.
- `source_code_url` : url where the step's source code can be viewed.
- `support_url` : url to the step's support / issue tracker.
- `published_at` : _auto-generated at share_ - step version's StepLib publish date
- `source` : _auto-generated at share_ git clone information.
- `asset_urls` : _auto-generated at share_ step assets (StepLib specific), like icon image.
- `host_os_tags` : supported operating systems. _Currently unused, reserved for future use._
- `project_type_tags` : project type tags if the step is project type specific.
  Example: `ios` or `android`. Completely optional, and only used for search
  and filtering in step lists.
- `type_tags` : generic type tags related to the step.
  Example: `utility`, `test` or `notification`.
  Similar to `project_type_tags`, this property is completely optional, and only used for search
  and filtering in step lists.
- `dependencies` : __DEPRECATED__ step dependency declarations.
- `deps` : the new, recommended step dependency declarations property.
- `toolkit` : step toolkit declaration, if the step is meant to utilize
  a Bitrise CLI provided toolkit (e.g. `Go`). If not defined the `Bash`
  toolkit is used by default.
- `is_requires_admin_user` : indication whether the step (might)
  require administrator rights for proper execution.
  _Currently unused, reserved for future use or will be deprecated (undecided)._
- `is_always_run` : if `true` the step will be executed even if a previous step failed during the build.
  Default is `false`.
- `is_skippable` : if `true`, even if the step fails that won't mark the build as failed,
  the error will be ignored. Default is `false`.
- `run_if` : a template based expression to declare when the step should run.
  If the expression evaluates to `true` the step will run, otherwise it will not.
  The default is a constant `true`.
    - `run_if: false` disables the step, and the step will always be skipped.
    - `run_if: .IsCI` will only run the step if the CLI runs in `CI` mode.
    - `run_if: '{{enveq "TEST_KEY" "test value"}}'` will skip the step unless
      the `TEST_KEY` environment variable is defined, and its value is `test value`.
- `inputs` : inputs Environments of the step.
- `outputs` : outputs Environments of the step.

## Environment properties

- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `is_expand` : if `true` the shell environment variables, in the Environment value, are expanded/resolved.
- `skip_if_empty` : if `true` and if the Environment's value is empty, these Environment will not be used.
- `category` : used to categorise the Environment variable.
- `value_options` : list of the available values.
- `is_required` : used when the Environment is used as a Step input Environment. If `true` the step requires to define not empty value for this Environment.
- `is_dont_change_value` : means, that this value should not be changed.
- `is_template` : if `true` the Environment's value will be evaulated as a go template and the evaulated value will be used.

