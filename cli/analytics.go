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
	sideEffects, err := env.GetDeclarationsSideEffects(environment, &env.DefaultEnvironmentSource{})
	if err != nil {
		return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
	}

	// Filter inputs from enviroments
	expandedInputs := make(map[string]string)
	for _, input := range inputs {
		inputKey, _, err := input.GetKeyValuePair()
		if err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}

		if sideEffects.ResultEnvironment[inputKey].IsSensitive {
			expandedInputs[inputKey] = filterwriter.RedactStr
			continue
		}

		src := bytes.NewReader([]byte(sideEffects.ResultEnvironment[inputKey].Value))
		dstBuf := new(bytes.Buffer)
		secretFilterDst := filterwriter.New(secrets, dstBuf)

		if _, err := io.Copy(secretFilterDst, src); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}
		if _, err := secretFilterDst.Flush(); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}

		expandedInputs[inputKey] = dstBuf.String()
	}

	return expandedInputs, nil
}
