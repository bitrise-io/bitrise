package cli

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise/tools/filterwriter"
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

		src := bytes.NewReader([]byte(inputValue))
		dstBuf := new(bytes.Buffer)
		secretFilterDst := filterwriter.New(secrets, dstBuf)

		if _, err := io.Copy(secretFilterDst, src); err != nil {
			return map[string]string{}, fmt.Errorf("failed to redact secrets, stream copy failed: %s", err)
		}
		if _, err := secretFilterDst.Flush(); err != nil {
			return map[string]string{}, fmt.Errorf("failed to redact secrets, stream flush failed: %s", err)
		}

		redactedStepInputs[inputKey] = dstBuf.String()
	}

	return redactedStepInputs, nil
}
