# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## StageModel Type

`object` ([Details](bitrise-defs-stagemodel.md))

# StageModel Properties

| Property                                  | Type      | Required | Nullable       | Defined by                                                                                                                                                                                  |
| :---------------------------------------- | :-------- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [title](#title)                           | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/title")                         |
| [summary](#summary)                       | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/summary")                     |
| [description](#description)               | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/description")             |
| [should\_always\_run](#should_always_run) | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/should_always_run") |
| [abort\_on\_fail](#abort_on_fail)         | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/abort_on_fail")         |
| [run\_if](#run_if)                        | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/run_if")                       |
| [workflows](#workflows)                   | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-stagemodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/workflows")                 |

## title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/title")

### title Type

`string`

## summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/summary")

### summary Type

`string`

## description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/description")

### description Type

`string`

## should\_always\_run



`should_always_run`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/should_always_run")

### should\_always\_run Type

`boolean`

## abort\_on\_fail



`abort_on_fail`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/abort_on_fail")

### abort\_on\_fail Type

`boolean`

## run\_if



`run_if`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-run_if.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/run_if")

### run\_if Type

`string`

## workflows



`workflows`

* is optional

* Type: `object[]` ([Details](bitrise-defs-stageworkflowlistitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stagemodel-properties-workflows.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StageModel/properties/workflows")

### workflows Type

`object[]` ([Details](bitrise-defs-stageworkflowlistitemmodel.md))
