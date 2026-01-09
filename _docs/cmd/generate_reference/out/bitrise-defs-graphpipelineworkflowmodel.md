# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## GraphPipelineWorkflowModel Type

`object` ([Details](bitrise-defs-graphpipelineworkflowmodel.md))

# GraphPipelineWorkflowModel Properties

| Property                                  | Type      | Required | Nullable       | Defined by                                                                                                                                                                                                                  |
| :---------------------------------------- | :-------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [depends\_on](#depends_on)                | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-depends_on.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/depends_on")               |
| [abort\_on\_fail](#abort_on_fail)         | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/abort_on_fail")         |
| [run\_if](#run_if)                        | `object`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelinerunifmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/run_if")                                            |
| [should\_always\_run](#should_always_run) | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/should_always_run") |
| [uses](#uses)                             | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-uses.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/uses")                           |
| [inputs](#inputs)                         | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/inputs")                       |
| [parallel](#parallel)                     | `string`  | Optional | cannot be null | [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-parallel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/parallel")                   |

## depends\_on



`depends_on`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-depends_on.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/depends_on")

### depends\_on Type

`string[]`

## abort\_on\_fail



`abort_on_fail`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-abort_on_fail.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/abort_on_fail")

### abort\_on\_fail Type

`boolean`

## run\_if



`run_if`

* is optional

* Type: `object` ([Details](bitrise-defs-graphpipelinerunifmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelinerunifmodel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/run_if")

### run\_if Type

`object` ([Details](bitrise-defs-graphpipelinerunifmodel.md))

## should\_always\_run



`should_always_run`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-should_always_run.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/should_always_run")

### should\_always\_run Type

`string`

## uses



`uses`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-uses.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/uses")

### uses Type

`string`

## inputs



`inputs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-graphpipelineworkflowmodelinput.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/inputs")

### inputs Type

`object[]` ([Details](bitrise-defs-graphpipelineworkflowmodelinput.md))

## parallel



`parallel`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-graphpipelineworkflowmodel-properties-parallel.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/GraphPipelineWorkflowModel/properties/parallel")

### parallel Type

`string`
