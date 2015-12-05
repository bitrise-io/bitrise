# Step Development Guideline

## Never depend on Environment Variables in your Step

You should expose every outside variable as an input of your step,
and just set the default value to the Environment Variable you want to use in the `step.yml`.

An example:

The Xcode Archive step generates a `$BITRISE_IPA_PATH` output environment variable.
**You should not** use this environment variable in your Step's code directly,
instead you should declare an input for your Step in `step.yml` and just set the default
value to `$BITRISE_IPA_PATH`. Example:

```
- ipa_path: "$BITRISE_IPA_PATH"
  opts:
      title: "IPA path"
```

After this, in your Step's code you can expect that the `$ipa_path` Environment Variable will
contain the value of the IPA path.

By declaring every option as an input you make it easier to test your Step,
and you also let the user of your Step to easily declare these inputs,
instead of searching in the code for the required Environment Variable.
