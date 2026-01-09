# Untitled object in undefined Schema

```txt
https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container
```



| Abstract            | Extensible | Status         | Identifiable | Custom Properties | Additional Properties | Access Restrictions | Defined In                                                          |
| :------------------ | :--------- | :------------- | :----------- | :---------------- | :-------------------- | :------------------ | :------------------------------------------------------------------ |
| Can be instantiated | No         | Unknown status | No           | Forbidden         | Forbidden             | none                | [bitrise.schema.json\*](bitrise.schema.json "open original schema") |

## Container Type

`object` ([Details](bitrise-defs-container.md))

# Container Properties

| Property                    | Type     | Required | Nullable       | Defined by                                                                                                                                                            |
| :-------------------------- | :------- | :------- | :------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [image](#image)             | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-image.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/image")     |
| [credentials](#credentials) | `object` | Optional | cannot be null | [Untitled schema](bitrise-defs-dockercredentials.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/credentials")        |
| [ports](#ports)             | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-ports.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/ports")     |
| [envs](#envs)               | `array`  | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/envs")       |
| [options](#options)         | `string` | Optional | cannot be null | [Untitled schema](bitrise-defs-container-properties-options.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/options") |

## image



`image`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-image.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/image")

### image Type

`string`

## credentials



`credentials`

* is optional

* Type: `object` ([Details](bitrise-defs-dockercredentials.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-dockercredentials.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/credentials")

### credentials Type

`object` ([Details](bitrise-defs-dockercredentials.md))

## ports



`ports`

* is optional

* Type: `string[]`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-ports.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/ports")

### ports Type

`string[]`

## envs



`envs`

* is optional

* Type: `object[]` ([Details](bitrise-defs-environmentitemmodel.md))

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-envs.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/envs")

### envs Type

`object[]` ([Details](bitrise-defs-environmentitemmodel.md))

## options



`options`

* is optional

* Type: `string`

* cannot be null

* defined in: [Untitled schema](bitrise-defs-container-properties-options.md "https://github.com/bitrise-io/bitrise/models/bitrise-data-model#/$defs/Container/properties/options")

### options Type

`string`
