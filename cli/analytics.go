package cli

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise/stepruncmd/filterwriter"
	"github.com/bitrise-io/envman/models"
)

func redactStepInputs(environment map[string]string, inputs []models.EnvironmentItemModel, secrets []string) (map[string]string, map[string]string, error) {
	redactedStepInputs := make(map[string]string)
	redactedOriginalInputs := map[string]string{}

	// Filter inputs from enviroments
	for _, input := range inputs {
		inputKey, originalValue, err := input.GetKeyValuePair()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get input key: %s", err)
		}

		// If input key may not be present in the result environment.
		// This can happen if the input has the "skip_if_empty" property set to true, and it is empty.
		inputValue, ok := environment[inputKey]
		if !ok {
			redactedStepInputs[inputKey] = ""
			continue
		}

		redactedValue, err := redactWithSecrets(inputValue, secrets)
		if err != nil {
			return nil, nil, err
		}
		redactedStepInputs[inputKey] = redactedValue

		if originalValue != "" && originalValue != inputValue {
			redactedOriginalValue, err := redactWithSecrets(originalValue, secrets)
			if err != nil {
				return nil, nil, err
			}
			redactedOriginalInputs[inputKey] = redactedOriginalValue
		}
	}

	return redactedStepInputs, redactedOriginalInputs, nil
}

func redactWithSecrets(inputValue string, secrets []string) (string, error) {
	src := bytes.NewReader([]byte(inputValue))
	dstBuf := new(bytes.Buffer)
	secretFilterDst := filterwriter.New(secrets, dstBuf)

	if _, err := io.Copy(secretFilterDst, src); err != nil {
		return "", fmt.Errorf("failed to redact secrets, stream copy failed: %s", err)
	}
	if err := secretFilterDst.Close(); err != nil {
		return "", fmt.Errorf("failed to redact secrets, closing the stream failed: %s", err)
	}

	redactedValue := dstBuf.String()
	return redactedValue, nil
}
