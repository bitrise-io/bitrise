# Untitled schema Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model
```



| Abstract            | Extensible | Status         | Identifiable            | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                        |
| :------------------ | :--------- | :------------- | :---------------------- | :---------------- | :-------------------- | :------------------ | :---------------------------------------------------------------- |
| Can be instantiated | No         | Unknown status | Unknown identifiability | Forbidden         | Allowed               | none                | [bitrise.schema.json](bitrise.schema.json "open original schema") |

## Untitled schema Type

unknown

# Untitled schema Definitions

## Definitions group AppModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel"}
```

| Property                                    | Type     | Required | Nullable       | Defined by                                                                                                                                                                                |
| :------------------------------------------ | :------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title)                             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/title")                           |
| [summary](#summary)                         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/summary")                       |
| [description](#description)                 | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/description")               |
| [status\_report\_name](#status_report_name) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/status_report_name") |
| [envs](#envs)                               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/envs")                             |

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/description")

#### description Type

`string`

### status\_report\_name



`status_report_name`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/status_report_name")

#### status\_report\_name Type

`string`

### envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/AppModel/properties/envs")

#### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

## Definitions group BitriseDataModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel"}
```

| Property                                               | Type     | Required | Nullable       | Defined by                                                                                                                                                                                                          |
| :----------------------------------------------------- | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [format\_version](#format_version)                     | `string` | Required | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-format_version.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/format_version")                   |
| [default\_step\_lib\_source](#default_step_lib_source) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-default_step_lib_source.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/default_step_lib_source") |
| [project\_type](#project_type)                         | `string` | Required | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-project_type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/project_type")                       |
| [title](#title-1)                                      | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/title")                                     |
| [summary](#summary-1)                                  | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/summary")                                 |
| [description](#description-1)                          | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/description")                         |
| [services](#services)                                  | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-services.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/services")                               |
| [containers](#containers)                              | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-containers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/containers")                           |
| [app](#app)                                            | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/app")                                                                |
| [meta](#meta)                                          | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/meta")                                       |
| [trigger\_map](#trigger_map)                           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/trigger_map")                                                 |
| [pipelines](#pipelines)                                | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-pipelines.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/pipelines")                             |
| [stages](#stages)                                      | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/stages")                                   |
| [workflows](#workflows)                                | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/workflows")                             |
| [step\_bundles](#step_bundles)                         | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-step_bundles.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/step_bundles")                       |

### format\_version



`format_version`

* is required

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-format_version.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/format_version")

#### format\_version Type

`string`

### default\_step\_lib\_source



`default_step_lib_source`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-default_step_lib_source.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/default_step_lib_source")

#### default\_step\_lib\_source Type

`string`

### project\_type



`project_type`

* is required

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-project_type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/project_type")

#### project\_type Type

`string`

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/description")

#### description Type

`string`

### services



`services`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-services.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-services.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/services")

#### services Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-services.md))

### containers



`containers`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-containers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-containers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/containers")

#### containers Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-containers.md))

### app



`app`

* is optional

* Type: `object` ([Details](bitrise-defs-appmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/app")

#### app Type

`object` ([Details](bitrise-defs-appmodel.md))

### meta



`meta`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-meta.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/meta")

#### meta Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-meta.md))

### trigger\_map



`trigger_map`

* is optional

* Type: `object[]` ([Details](bitrise-defs-triggermapitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/trigger_map")

#### trigger\_map Type

`object[]` ([Details](bitrise-defs-triggermapitemmodel.md))

### pipelines



`pipelines`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-pipelines.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-pipelines.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/pipelines")

#### pipelines Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-pipelines.md))

### stages



`stages`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-stages.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/stages")

#### stages Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-stages.md))

### workflows



`workflows`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-workflows.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/workflows")

#### workflows Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-workflows.md))

### step\_bundles



`step_bundles`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-step_bundles.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-step_bundles.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/step_bundles")

#### step\_bundles Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-step_bundles.md))

## Definitions group Container

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container"}
```

| Property                    | Type     | Required | Nullable       | Defined by                                                                                                                                                            |
| :-------------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [image](#image)             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-image.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/image")     |
| [credentials](#credentials) | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-dockercredentials.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/credentials")        |
| [ports](#ports)             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-ports.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/ports")     |
| [envs](#envs-1)             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/envs")       |
| [options](#options)         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-options.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/options") |

### image



`image`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-image.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/image")

#### image Type

`string`

### credentials



`credentials`

* is optional

* Type: `object` ([Details](bitrise-defs-dockercredentials.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-dockercredentials.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/credentials")

#### credentials Type

`object` ([Details](bitrise-defs-dockercredentials.md))

### ports



`ports`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-ports.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/ports")

#### ports Type

`string[]`

### envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/envs")

#### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

### options



`options`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-options.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/options")

#### options Type

`string`

## Definitions group DockerCredentials

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials"}
```

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                                              |
| :-------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [username](#username) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-dockercredentials-properties-username.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/username") |
| [password](#password) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-dockercredentials-properties-password.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/password") |
| [server](#server)     | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-dockercredentials-properties-server.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/server")     |

### username



`username`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-dockercredentials-properties-username.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/username")

#### username Type

`string`

### password



`password`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-dockercredentials-properties-password.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/password")

#### password Type

`string`

### server



`server`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-dockercredentials-properties-server.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/DockerCredentials/properties/server")

#### server Type

`string`

## Definitions group EnvironmentItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/EnvironmentItemModel"}
```

| Property | Type | Required | Nullable | Defined by |
| :------- | :--- | :------- | :------- | :--------- |

## Definitions group GraphPipelineRunIfModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineRunIfModel"}
```

| Property                  | Type     | Required | Nullable       | Defined by                                                                                                                                                                                              |
| :------------------------ | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [expression](#expression) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelinerunifmodel-properties-expression.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineRunIfModel/properties/expression") |

### expression



`expression`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelinerunifmodel-properties-expression.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineRunIfModel/properties/expression")

#### expression Type

`string`

## Definitions group GraphPipelineWorkflowListItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowListItemModel"}
```

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                                                     |
| :-------------------- | :------- | :------- | :------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Additional Properties | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowListItemModel/additionalProperties") |

### Additional Properties

Additional properties are allowed, as long as they follow this schema:



* is optional

* Type: `object` ([Details](bitrise-defs-graphpipelineworkflowmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowListItemModel/additionalProperties")

#### additionalProperties Type

`object` ([Details](bitrise-defs-graphpipelineworkflowmodel.md))

## Definitions group GraphPipelineWorkflowModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel"}
```

| Property                                  | Type      | Required | Nullable       | Defined by                                                                                                                                                                                                                  |
| :---------------------------------------- | :-------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [depends\_on](#depends_on)                | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-depends_on.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/depends_on")               |
| [abort\_on\_fail](#abort_on_fail)         | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/abort_on_fail")         |
| [run\_if](#run_if)                        | `object`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelinerunifmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/run_if")                                            |
| [should\_always\_run](#should_always_run) | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/should_always_run") |
| [uses](#uses)                             | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-uses.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/uses")                           |
| [inputs](#inputs)                         | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/inputs")                       |
| [parallel](#parallel)                     | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-parallel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/parallel")                   |

### depends\_on



`depends_on`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-depends_on.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/depends_on")

#### depends\_on Type

`string[]`

### abort\_on\_fail



`abort_on_fail`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/abort_on_fail")

#### abort\_on\_fail Type

`boolean`

### run\_if



`run_if`

* is optional

* Type: `object` ([Details](bitrise-defs-graphpipelinerunifmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelinerunifmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/run_if")

#### run\_if Type

`object` ([Details](bitrise-defs-graphpipelinerunifmodel.md))

### should\_always\_run



`should_always_run`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/should_always_run")

#### should\_always\_run Type

`string`

### uses



`uses`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-uses.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/uses")

#### uses Type

`string`

### inputs



`inputs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-graphpipelineworkflowmodelinput.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/inputs")

#### inputs Type

`object[]` ([Details](bitrise-defs-graphpipelineworkflowmodelinput.md))

### parallel



`parallel`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-parallel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/parallel")

#### parallel Type

`string`

## Definitions group GraphPipelineWorkflowModelInput

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModelInput"}
```

| Property | Type | Required | Nullable | Defined by |
| :------- | :--- | :------- | :------- | :--------- |

## Definitions group PipelineModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel"}
```

| Property                                      | Type     | Required | Nullable       | Defined by                                                                                                                                                                                          |
| :-------------------------------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title-2)                             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/title")                           |
| [summary](#summary-2)                         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/summary")                       |
| [description](#description-2)                 | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/description")               |
| [triggers](#triggers)                         | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/triggers")                                              |
| [status\_report\_name](#status_report_name-1) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/status_report_name") |
| [stages](#stages-1)                           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/stages")                         |
| [workflows](#workflows-1)                     | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowlistitemmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/workflows")                   |

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/description")

#### description Type

`string`

### triggers



`triggers`

* is optional

* Type: `object` ([Details](bitrise-defs-triggers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/triggers")

#### triggers Type

`object` ([Details](bitrise-defs-triggers.md))

### status\_report\_name



`status_report_name`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/status_report_name")

#### status\_report\_name Type

`string`

### stages



`stages`

* is optional

* Type: `object[]` ([Details](bitrise-defs-stagelistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/stages")

#### stages Type

`object[]` ([Details](bitrise-defs-stagelistitemmodel.md))

### workflows



`workflows`

* is optional

* Type: `object` ([Details](bitrise-defs-graphpipelineworkflowlistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowlistitemmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/workflows")

#### workflows Type

`object` ([Details](bitrise-defs-graphpipelineworkflowlistitemmodel.md))

## Definitions group PullRequestGitEventTriggerItem

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PullRequestGitEventTriggerItem"}
```

| Property                           | Type          | Required | Nullable       | Defined by                                                                                                                                                                                                                  |
| :--------------------------------- | :------------ | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [enabled](#enabled)                | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-pullrequestgiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PullRequestGitEventTriggerItem/properties/enabled")             |
| [draft\_enabled](#draft_enabled)   | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-pullrequestgiteventtriggeritem-properties-draft_enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PullRequestGitEventTriggerItem/properties/draft_enabled") |
| [source\_branch](#source_branch)   | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |
| [target\_branch](#target_branch)   | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |
| [label](#label)                    | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |
| [comment](#comment)                | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |
| [commit\_message](#commit_message) | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |
| [changed\_files](#changed_files)   | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                       |

### enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pullrequestgiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PullRequestGitEventTriggerItem/properties/enabled")

#### enabled Type

`boolean`

### draft\_enabled



`draft_enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pullrequestgiteventtriggeritem-properties-draft_enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PullRequestGitEventTriggerItem/properties/draft_enabled")

#### draft\_enabled Type

`boolean`

### source\_branch

no description

`source_branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### target\_branch

no description

`target_branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### label

no description

`label`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### comment

no description

`comment`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### commit\_message

no description

`commit_message`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### changed\_files

no description

`changed_files`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

## Definitions group PushGitEventTriggerItem

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PushGitEventTriggerItem"}
```

| Property                             | Type          | Required | Nullable       | Defined by                                                                                                                                                                                        |
| :----------------------------------- | :------------ | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [enabled](#enabled-1)                | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-pushgiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PushGitEventTriggerItem/properties/enabled") |
| [branch](#branch)                    | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                             |
| [commit\_message](#commit_message-1) | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                             |
| [changed\_files](#changed_files-1)   | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                             |

### enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pushgiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PushGitEventTriggerItem/properties/enabled")

#### enabled Type

`boolean`

### branch

no description

`branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### commit\_message

no description

`commit_message`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### changed\_files

no description

`changed_files`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

## Definitions group StageListItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageListItemModel"}
```

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                     |
| :-------------------- | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Additional Properties | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageListItemModel/additionalProperties") |

### Additional Properties

Additional properties are allowed, as long as they follow this schema:



* is optional

* Type: `object` ([Details](bitrise-defs-stagemodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageListItemModel/additionalProperties")

#### additionalProperties Type

`object` ([Details](bitrise-defs-stagemodel.md))

## Definitions group StageModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel"}
```

| Property                                    | Type      | Required | Nullable       | Defined by                                                                                                                                                                                  |
| :------------------------------------------ | :-------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [title](#title-3)                           | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/title")                         |
| [summary](#summary-3)                       | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/summary")                     |
| [description](#description-3)               | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/description")             |
| [should\_always\_run](#should_always_run-1) | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/should_always_run") |
| [abort\_on\_fail](#abort_on_fail-1)         | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/abort_on_fail")         |
| [run\_if](#run_if-1)                        | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/run_if")                       |
| [workflows](#workflows-2)                   | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/workflows")                 |

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/description")

#### description Type

`string`

### should\_always\_run



`should_always_run`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/should_always_run")

#### should\_always\_run Type

`boolean`

### abort\_on\_fail



`abort_on_fail`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/abort_on_fail")

#### abort\_on\_fail Type

`boolean`

### run\_if



`run_if`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/run_if")

#### run\_if Type

`string`

### workflows



`workflows`

* is optional

* Type: `object[]` ([Details](bitrise-defs-stageworkflowlistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/workflows")

#### workflows Type

`object[]` ([Details](bitrise-defs-stageworkflowlistitemmodel.md))

## Definitions group StageWorkflowListItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowListItemModel"}
```

| Property              | Type     | Required | Nullable       | Defined by                                                                                                                                                                     |
| :-------------------- | :------- | :------- | :------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Additional Properties | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-stageworkflowmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowListItemModel/additionalProperties") |

### Additional Properties

Additional properties are allowed, as long as they follow this schema:



* is optional

* Type: `object` ([Details](bitrise-defs-stageworkflowmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stageworkflowmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowListItemModel/additionalProperties")

#### additionalProperties Type

`object` ([Details](bitrise-defs-stageworkflowmodel.md))

## Definitions group StageWorkflowModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowModel"}
```

| Property             | Type     | Required | Nullable       | Defined by                                                                                                                                                                            |
| :------------------- | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [run\_if](#run_if-2) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stageworkflowmodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowModel/properties/run_if") |

### run\_if



`run_if`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stageworkflowmodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageWorkflowModel/properties/run_if")

#### run\_if Type

`string`

## Definitions group StepBundleModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel"}
```

| Property                      | Type     | Required | Nullable       | Defined by                                                                                                                                                                                |
| :---------------------------- | :------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title-4)             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/title")             |
| [summary](#summary-4)         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/summary")         |
| [description](#description-4) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/description") |
| [inputs](#inputs-1)           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/inputs")           |
| [envs](#envs-2)               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/envs")               |
| [steps](#steps)               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/steps")             |

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/description")

#### description Type

`string`

### inputs



`inputs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/inputs")

#### inputs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

### envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/envs")

#### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

### steps



`steps`

* is optional

* Type: `object[]` ([Details](bitrise-defs-steplistitemsteporbundlemodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/steps")

#### steps Type

`object[]` ([Details](bitrise-defs-steplistitemsteporbundlemodel.md))

## Definitions group StepListItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepListItemModel"}
```

| Property | Type | Required | Nullable | Defined by |
| :------- | :--- | :------- | :------- | :--------- |

## Definitions group StepListItemStepOrBundleModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepListItemStepOrBundleModel"}
```

| Property | Type | Required | Nullable | Defined by |
| :------- | :--- | :------- | :------- | :--------- |

## Definitions group TagGitEventTriggerItem

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TagGitEventTriggerItem"}
```

| Property              | Type          | Required | Nullable       | Defined by                                                                                                                                                                                      |
| :-------------------- | :------------ | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [enabled](#enabled-2) | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-taggiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TagGitEventTriggerItem/properties/enabled") |
| [name](#name)         | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                           |

### enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-taggiteventtriggeritem-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TagGitEventTriggerItem/properties/enabled")

#### enabled Type

`boolean`

### name

no description

`name`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

## Definitions group TriggerMapItemModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel"}
```

| Property                                                     | Type          | Required | Nullable       | Defined by                                                                                                                                                                                                                      |
| :----------------------------------------------------------- | :------------ | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [type](#type)                                                | `string`      | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/type")                                             |
| [enabled](#enabled-3)                                        | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/enabled")                                       |
| [pipeline](#pipeline)                                        | `string`      | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-pipeline.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/pipeline")                                     |
| [workflow](#workflow)                                        | `string`      | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-workflow.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/workflow")                                     |
| [push\_branch](#push_branch)                                 | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [commit\_message](#commit_message-2)                         | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [changed\_files](#changed_files-2)                           | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [tag](#tag)                                                  | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [pull\_request\_source\_branch](#pull_request_source_branch) | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [pull\_request\_target\_branch](#pull_request_target_branch) | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [draft\_pull\_request\_enabled](#draft_pull_request_enabled) | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-draft_pull_request_enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/draft_pull_request_enabled") |
| [pull\_request\_label](#pull_request_label)                  | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [pull\_request\_comment](#pull_request_comment)              | Not specified | Optional | cannot be null | [Untitled schema](undefined.md "undefined#undefined")                                                                                                                                                                           |
| [pattern](#pattern)                                          | `string`      | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-pattern.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/pattern")                                       |
| [is\_pull\_request\_allowed](#is_pull_request_allowed)       | `boolean`     | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapitemmodel-properties-is_pull_request_allowed.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/is_pull_request_allowed")       |

### type



`type`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/type")

#### type Type

`string`

### enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/enabled")

#### enabled Type

`boolean`

### pipeline



`pipeline`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-pipeline.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/pipeline")

#### pipeline Type

`string`

### workflow



`workflow`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-workflow.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/workflow")

#### workflow Type

`string`

### push\_branch

no description

`push_branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### commit\_message

no description

`commit_message`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### changed\_files

no description

`changed_files`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### tag

no description

`tag`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### pull\_request\_source\_branch

no description

`pull_request_source_branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### pull\_request\_target\_branch

no description

`pull_request_target_branch`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### draft\_pull\_request\_enabled



`draft_pull_request_enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-draft_pull_request_enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/draft_pull_request_enabled")

#### draft\_pull\_request\_enabled Type

`boolean`

### pull\_request\_label

no description

`pull_request_label`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### pull\_request\_comment

no description

`pull_request_comment`

* is optional

* Type: unknown

* cannot be null

* defined in: [Untitled schema](undefined.md "undefined#undefined")

#### Untitled schema Type

unknown

### pattern



`pattern`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-pattern.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/pattern")

#### pattern Type

`string`

### is\_pull\_request\_allowed



`is_pull_request_allowed`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapitemmodel-properties-is_pull_request_allowed.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapItemModel/properties/is_pull_request_allowed")

#### is\_pull\_request\_allowed Type

`boolean`

## Definitions group TriggerMapModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/TriggerMapModel"}
```

| Property | Type | Required | Nullable | Defined by |
| :------- | :--- | :------- | :------- | :--------- |

## Definitions group Triggers

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers"}
```

| Property                       | Type      | Required | Nullable       | Defined by                                                                                                                                                                    |
| :----------------------------- | :-------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [enabled](#enabled-4)          | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/enabled")           |
| [push](#push)                  | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-push.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/push")                 |
| [pull\_request](#pull_request) | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-pull_request.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/pull_request") |
| [tag](#tag-1)                  | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-tag.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/tag")                   |

### enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/enabled")

#### enabled Type

`boolean`

### push



`push`

* is optional

* Type: `object[]` ([Details](bitrise-defs-pushgiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-push.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/push")

#### push Type

`object[]` ([Details](bitrise-defs-pushgiteventtriggeritem.md))

### pull\_request



`pull_request`

* is optional

* Type: `object[]` ([Details](bitrise-defs-pullrequestgiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-pull_request.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/pull_request")

#### pull\_request Type

`object[]` ([Details](bitrise-defs-pullrequestgiteventtriggeritem.md))

### tag



`tag`

* is optional

* Type: `object[]` ([Details](bitrise-defs-taggiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-tag.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/tag")

#### tag Type

`object[]` ([Details](bitrise-defs-taggiteventtriggeritem.md))

## Definitions group WorkflowModel

Reference this group by using

```json
{"$ref":"https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel"}
```

| Property                                      | Type     | Required | Nullable       | Defined by                                                                                                                                                                                          |
| :-------------------------------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title-5)                             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/title")                           |
| [summary](#summary-5)                         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/summary")                       |
| [description](#description-5)                 | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/description")               |
| [triggers](#triggers-1)                       | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/triggers")                                              |
| [status\_report\_name](#status_report_name-2) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/status_report_name") |
| [before\_run](#before_run)                    | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-before_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/before_run")                 |
| [after\_run](#after_run)                      | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-after_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/after_run")                   |
| [envs](#envs-3)                               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/envs")                             |
| [steps](#steps-1)                             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/steps")                           |
| [meta](#meta-1)                               | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/meta")                             |

### title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/title")

#### title Type

`string`

### summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/summary")

#### summary Type

`string`

### description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/description")

#### description Type

`string`

### triggers



`triggers`

* is optional

* Type: `object` ([Details](bitrise-defs-triggers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/triggers")

#### triggers Type

`object` ([Details](bitrise-defs-triggers.md))

### status\_report\_name



`status_report_name`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/status_report_name")

#### status\_report\_name Type

`string`

### before\_run



`before_run`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-before_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/before_run")

#### before\_run Type

`string[]`

### after\_run



`after_run`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-after_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/after_run")

#### after\_run Type

`string[]`

### envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/envs")

#### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

### steps



`steps`

* is optional

* Type: `object[]` ([Details](bitrise-defs-steplistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/steps")

#### steps Type

`object[]` ([Details](bitrise-defs-steplistitemmodel.md))

### meta



`meta`

* is optional

* Type: `object` ([Details](bitrise-defs-workflowmodel-properties-meta.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/meta")

#### meta Type

`object` ([Details](bitrise-defs-workflowmodel-properties-meta.md))
