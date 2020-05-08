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
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}

		src := bytes.NewReader([]byte(environment[inputKey]))
		dstBuf := new(bytes.Buffer)
		secretFilterDst := filterwriter.New(secrets, dstBuf)

		if _, err := io.Copy(secretFilterDst, src); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}
		if _, err := secretFilterDst.Flush(); err != nil {
			return map[string]string{}, fmt.Errorf("expandStepInputsForAnalytics() failed, %s", err)
		}

		redactStepInputs[inputKey] = dstBuf.String()
	}

	return redactStepInputs, nil
}
