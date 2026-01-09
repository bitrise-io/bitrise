# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## WorkflowModel Type

`object` ([Details](bitrise-defs-workflowmodel.md))

# WorkflowModel Properties

| Property                                    | Type     | Required | Nullable       | Defined by                                                                                                                                                                                          |
| :------------------------------------------ | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title)                             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/title")                           |
| [summary](#summary)                         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/summary")                       |
| [description](#description)                 | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/description")               |
| [triggers](#triggers)                       | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/triggers")                                              |
| [status\_report\_name](#status_report_name) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/status_report_name") |
| [before\_run](#before_run)                  | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-before_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/before_run")                 |
| [after\_run](#after_run)                    | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-after_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/after_run")                   |
| [envs](#envs)                               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/envs")                             |
| [steps](#steps)                             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/steps")                           |
| [meta](#meta)                               | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-workflowmodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/meta")                             |

## title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/title")

### title Type

`string`

## summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/summary")

### summary Type

`string`

## description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/description")

### description Type

`string`

## triggers



`triggers`

* is optional

* Type: `object` ([Details](bitrise-defs-triggers.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/triggers")

### triggers Type

`object` ([Details](bitrise-defs-triggers.md))

## status\_report\_name



`status_report_name`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-status_report_name.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/status_report_name")

### status\_report\_name Type

`string`

## before\_run



`before_run`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-before_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/before_run")

### before\_run Type

`string[]`

## after\_run



`after_run`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-after_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/after_run")

### after\_run Type

`string[]`

## envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/envs")

### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

## steps



`steps`

* is optional

* Type: `object[]` ([Details](bitrise-defs-steplistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/steps")

### steps Type

`object[]` ([Details](bitrise-defs-steplistitemmodel.md))

## meta



`meta`

* is optional

* Type: `object` ([Details](bitrise-defs-workflowmodel-properties-meta.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-workflowmodel-properties-meta.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/WorkflowModel/properties/meta")

### meta Type

`object` ([Details](bitrise-defs-workflowmodel-properties-meta.md))
