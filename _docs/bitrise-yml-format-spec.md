## Minimal bitrise.yml

The bare minimum `bitrise.yml` is:

```
format_version: 1.3.1
```

Minimum `bitrise.yml` for a single (no-op) workflow:

```
format_version: 1.3.1
workflows:
  test:
```

## Top level bitrise.yml properties

- `format_version` : this property declares the minimum Bitrise CLI format version.
  You can get your Bitrise CLI's supported highest format version with: `bitrise version --full`.
  If you set the `format_version` to `1.3.1` that means that Bitrise CLI versions which
  don't support the format version `1.3.1` or higher won't be able to run the configuration.
  This is important if you use features which are not available in older Bitrise CLI versions.
- `default_step_lib_source` : specifies the source to use when no other source is defined for a step.
- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `app` : global, "app" specific configurations. The following optional sub properties are supported:
    - `envs` : configuration global environment variables list
    - `title`, `summary` and `description` : metadata, for comments, tools and GUI.
      _Note: these meta properties can be used for permanent comments. Standard YML comments
      are not preserved when the YML is normalized, converted to JSON or otherwise
      generated or transformed. These meta properties are._
- `trigger_map` : Trigger Map definitions.
- `workflows` : workflow definitions.


## Workflow properties

- `title`, `summary` and `description` : metadata, for comments, tools and GUI.
  _Note: these meta properties can be used for permanent comments. Standard YML comments
  are not preserved when the YML is normalized, converted to JSON or otherwise
  generated or transformed. These meta properties are._
- `before_run` : list of workflows to execute before this workflow
- `after_run` : list of workflows to execute after this workflow
- `envs` : workflow defined environment variables list
- `steps` : workflow defined step list
