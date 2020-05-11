package cli

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise/tools/filterwriter"
	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/models"
)

func expandStepInputsForAnalytics(environment, inputs []models.EnvironmentItemModel, secrets []string) (map[string]string, error) {
	for _, newEnv := range environment {
		if err := newEnv.FillMissingDefaults(); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, could not fill env (%s) defaults: %s", newEnv, err)
		}
	}

	sideEffects, err := env.GetDeclarationsSideEffects(environment, &env.DefaultEnvironmentSource{})
	if err != nil {
		return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
	}

	// Filter inputs from enviroments
	expandedInputs := make(map[string]string)
	for _, input := range inputs {
		inputKey, _, err := input.GetKeyValuePair()
		if err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to get input key: %s", err)
		}

		// If input key not found in the result environment, it will be an empty string.
		// This can happen if the input has the "skip_if_empty" property set to true, and it is empty.
		src := bytes.NewReader([]byte(sideEffects.ResultEnvironment[inputKey]))
		dstBuf := new(bytes.Buffer)
		secretFilterDst := filterwriter.New(secrets, dstBuf)

		if _, err := io.Copy(secretFilterDst, src); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to redact secrets, stream copy failure: %s", err)
		}
		if _, err := secretFilterDst.Flush(); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed to redact secrets, stream flush failure: %s", err)
		}

		expandedInputs[inputKey] = dstBuf.String()
	}

	return expandedInputs, nil
}
