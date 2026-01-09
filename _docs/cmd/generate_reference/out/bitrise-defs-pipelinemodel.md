# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## PipelineModel Type

`object` ([Details](bitrise-defs-pipelinemodel.md))

# PipelineModel Properties

| Property                                    | Type     | Required | Nullable       | Defined by                                                                                                                                                                                          |
| :------------------------------------------ | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title)                             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/title")                           |
| [summary](#summary)                         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/summary")                       |
| [description](#description)                 | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/description")               |
| [triggers](#triggers)                       | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/triggers")                                              |
| [status\_report\_name](#status_report_name) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/status_report_name") |
| [stages](#stages)                           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-pipelinemodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/stages")                         |
| [workflows](#workflows)                     | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowlistitemmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/workflows")                   |

## title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/title")

### title Type

`string`

## summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/summary")

### summary Type

`string`

## description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/description")

### description Type

`string`

## triggers



`triggers`

* is optional

* Type: `object` ([Details](bitrise-defs-triggers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/triggers")

### triggers Type

`object` ([Details](bitrise-defs-triggers.md))

## status\_report\_name



`status_report_name`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/status_report_name")

### status\_report\_name Type

`string`

## stages



`stages`

* is optional

* Type: `object[]` ([Details](bitrise-defs-stagelistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-pipelinemodel-properties-stages.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/stages")

### stages Type

`object[]` ([Details](bitrise-defs-stagelistitemmodel.md))

## workflows



`workflows`

* is optional

* Type: `object` ([Details](bitrise-defs-graphpipelineworkflowlistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowlistitemmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/PipelineModel/properties/workflows")

### workflows Type

`object` ([Details](bitrise-defs-graphpipelineworkflowlistitemmodel.md))
