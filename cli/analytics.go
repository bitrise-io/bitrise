package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/envman/models"
)

func redactStepInputs(environment map[string]string, inputs []models.EnvironmentItemModel, secrets []string) (map[string]string, error) {
	redactedStepInputs := make(map[string]string)

	// Filter inputs from enviroments
	for _, input := range inputs {
		inputKey, _, err := input.GetKeyValuePair()
		if err != nil {
			return map[string]string{}, fmt.Errorf("failed to get input key: %s", err)
		}

		// If input key may not be present in the result environment.
		// This can happen if the input has the "skip_if_empty" property set to true, and it is empty.
		inputValue, ok := environment[inputKey]
		if !ok {
			redactedStepInputs[inputKey] = ""
			continue
		}

		redacted, err := analytics.RedactStringWithSecret(inputValue, secrets)
		if err != nil {
			return map[string]string{}, err
		}

		redactedStepInputs[inputKey] = redacted
	}

	return redactedStepInputs, nil
}
