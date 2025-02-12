package integration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestStepBundleInputs(t *testing.T) {
	configPth := "step_bundles_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "test_step_bundle_inputs", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello Bitrise!")
}

func TestNestedStepBundle(t *testing.T) {
	configPth := "step_bundles_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "--output-format", "json", "test_nested_step_bundle", "-c", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	stepOutputs := collectStepOutputs(out, t)
	require.Equal(t, stepOutputs, []string{
		`bundle1
bundle1_input1: bundle1_input1
bundle2_input1: bundle2_input1
`,
		`bundle1
bundle1_input1: bundle1_input1_override
bundle2_input1: bundle2_input1
`,
		`bundle2
bundle1_input1: 
bundle2_input1: bundle2_input1
`,
		`workflow step
bundle1_input1: 
bundle2_input1: 
`,
	})
}

func collectStepOutputs(workflowRunOut string, t *testing.T) []string {
	type messageLine struct {
		Producer   string `json:"producer"`
		ProducerID string `json:"producer_id"`
		Message    string `json:"message"`
	}

	idxToStep := map[string]int{}
	messageToStep := map[string]string{}
	stepIDx := 0

	for _, line := range strings.Split(workflowRunOut, "\n") {
		var message messageLine
		err := json.Unmarshal([]byte(line), &message)
		require.NoError(t, err)

		if message.Producer != "step" {
			continue
		}

		stepMessage := messageToStep[message.ProducerID]
		stepMessage += message.Message
		messageToStep[message.ProducerID] = stepMessage

		_, ok := idxToStep[message.ProducerID]
		if !ok {
			idxToStep[message.ProducerID] = stepIDx
			stepIDx++
		}
	}

	stepMessages := make([]string, len(idxToStep))
	for step, idx := range idxToStep {
		stepMessages[idx] = messageToStep[step]
	}

	return stepMessages
}
