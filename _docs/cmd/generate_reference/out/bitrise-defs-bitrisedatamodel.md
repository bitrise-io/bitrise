# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## BitriseDataModel Type

`object` ([Details](bitrise-defs-bitrisedatamodel.md))

# BitriseDataModel Properties

| Property                                               | Type     | Required | Nullable       | Defined by                                                                                                                                                                                                          |
| :----------------------------------------------------- | :------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [format\_version](#format_version)                     | `string` | Required | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-format_version.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/format_version")                   |
| [default\_step\_lib\_source](#default_step_lib_source) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-default_step_lib_source.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/default_step_lib_source") |
| [project\_type](#project_type)                         | `string` | Required | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-project_type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/project_type")                       |
| [title](#title)                                        | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/title")                                     |
| [summary](#summary)                                    | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/summary")                                 |
| [description](#description)                            | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/description")                         |
| [services](#services)                                  | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-services.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/services")                               |
| [containers](#containers)                              | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-containers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/containers")                           |
| [app](#app)                                            | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-appmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/app")                                                                |
| [meta](#meta)                                          | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/meta")                                       |
| [trigger\_map](#trigger_map)                           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-triggermapmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/trigger_map")                                                 |
| [pipelines](#pipelines)                                | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-pipelines.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/pipelines")                             |
| [stages](#stages)                                      | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/stages")                                   |
| [workflows](#workflows)                                | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/workflows")                             |
| [step\_bundles](#step_bundles)                         | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-bitrisedatamodel-properties-step_bundles.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/step_bundles")                       |

## format\_version



`format_version`

* is required

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-format_version.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/format_version")

### format\_version Type

`string`

## default\_step\_lib\_source



`default_step_lib_source`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-default_step_lib_source.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/default_step_lib_source")

### default\_step\_lib\_source Type

`string`

## project\_type



`project_type`

* is required

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-project_type.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/project_type")

### project\_type Type

`string`

## title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/title")

### title Type

`string`

## summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/summary")

### summary Type

`string`

## description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/description")

### description Type

`string`

## services



`services`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-services.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-services.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/services")

### services Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-services.md))

## containers



`containers`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-containers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-containers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/containers")

### containers Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-containers.md))

## app



`app`

* is optional

* Type: `object` ([Details](bitrise-defs-appmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-appmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/app")

### app Type

`object` ([Details](bitrise-defs-appmodel.md))

## meta



`meta`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-meta.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/meta")

### meta Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-meta.md))

## trigger\_map



`trigger_map`

* is optional

* Type: `object[]` ([Details](bitrise-defs-triggermapitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggermapmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/trigger_map")

### trigger\_map Type

`object[]` ([Details](bitrise-defs-triggermapitemmodel.md))

## pipelines



`pipelines`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-pipelines.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-pipelines.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/pipelines")

### pipelines Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-pipelines.md))

## stages



`stages`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-stages.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/stages")

### stages Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-stages.md))

## workflows



`workflows`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-workflows.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/workflows")

### workflows Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-workflows.md))

## step\_bundles



`step_bundles`

* is optional

* Type: `object` ([Details](bitrise-defs-bitrisedatamodel-properties-step_bundles.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-bitrisedatamodel-properties-step_bundles.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/BitriseDataModel/properties/step_bundles")

### step\_bundles Type

`object` ([Details](bitrise-defs-bitrisedatamodel-properties-step_bundles.md))
