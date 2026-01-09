# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## Triggers Type

`object` ([Details](bitrise-defs-triggers.md))

# Triggers Properties

| Property                       | Type      | Required | Nullable       | Defined by                                                                                                                                                                    |
| :----------------------------- | :-------- | :------- | :------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [enabled](#enabled)            | `boolean` | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/enabled")           |
| [push](#push)                  | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-push.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/push")                 |
| [pull\_request](#pull_request) | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-pull_request.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/pull_request") |
| [tag](#tag)                    | `array`   | Optional | cannot be null | [Untitled schema](bitrise-defs-triggers-properties-tag.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/tag")                   |

## enabled



`enabled`

* is optional

* Type: `boolean`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-enabled.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/enabled")

### enabled Type

`boolean`

## push



`push`

* is optional

* Type: `object[]` ([Details](bitrise-defs-pushgiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-push.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/push")

### push Type

`object[]` ([Details](bitrise-defs-pushgiteventtriggeritem.md))

## pull\_request



`pull_request`

* is optional

* Type: `object[]` ([Details](bitrise-defs-pullrequestgiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-pull_request.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/pull_request")

### pull\_request Type

`object[]` ([Details](bitrise-defs-pullrequestgiteventtriggeritem.md))

## tag



`tag`

* is optional

* Type: `object[]` ([Details](bitrise-defs-taggiteventtriggeritem.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-triggers-properties-tag.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Triggers/properties/tag")

### tag Type

`object[]` ([Details](bitrise-defs-taggiteventtriggeritem.md))
