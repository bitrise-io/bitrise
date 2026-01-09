# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## StepBundleModel Type

`object` ([Details](bitrise-defs-stepbundlemodel.md))

# StepBundleModel Properties

| Property                    | Type     | Required | Nullable       | Defined by                                                                                                                                                                                |
| :-------------------------- | :------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [title](#title)             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/title")             |
| [summary](#summary)         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/summary")         |
| [description](#description) | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/description") |
| [inputs](#inputs)           | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/inputs")           |
| [envs](#envs)               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/envs")               |
| [steps](#steps)             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-stepbundlemodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/steps")             |

## title



`title`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-title.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/title")

### title Type

`string`

## summary



`summary`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-summary.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/summary")

### summary Type

`string`

## description



`description`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-description.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/description")

### description Type

`string`

## inputs



`inputs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-inputs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/inputs")

### inputs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

## envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/envs")

### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

## steps



`steps`

* is optional

* Type: `object[]` ([Details](bitrise-defs-steplistitemsteporbundlemodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-stepbundlemodel-properties-steps.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/StepBundleModel/properties/steps")

### steps Type

`object[]` ([Details](bitrise-defs-steplistitemsteporbundlemodel.md))
