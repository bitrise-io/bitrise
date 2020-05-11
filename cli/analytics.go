package cli

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise/tools/filterwriter"
	"github.com/bitrise-io/envman/models"
)

func redactStepInputs(environment map[string]string, inputs []models.EnvironmentItemModel, secrets []string) (map[string]string, error) {
	redactStepInputs := make(map[string]string)

	// Filter inputs from enviroments
	for _, input := range inputs {
		inputKey, _, err := input.GetKeyValuePair()
		if err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to get input key: %s", err)
		}

		// If input key not found in the result environment, it will be an empty string.
		// This can happen if the input has the "skip_if_empty" property set to true, and it is empty.
		src := bytes.NewReader([]byte(environment[inputKey]))
		dstBuf := new(bytes.Buffer)
		secretFilterDst := filterwriter.New(secrets, dstBuf)

		if _, err := io.Copy(secretFilterDst, src); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to redact secrets, stream copy failure: %s", err)
		}
		if _, err := secretFilterDst.Flush(); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to redact secrets, stream flush failure: %s", err)
		}

		redactStepInputs[inputKey] = dstBuf.String()
	}

	return redactStepInputs, nil
}
