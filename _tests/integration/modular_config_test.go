package integration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_ModularConfig(t *testing.T) {
	configPth := "modular_config_main.yml"
	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")

	cmd := command.New(binPath(), "merge", configPth, "-o", deployDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)

	cmd = command.New(binPath(), "validate", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Equal(t, "Config is valid: \u001B[32;1mtrue\u001B[0m", out)

	cmd = command.New(binPath(), "workflows", "--id-only", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Equal(t, "print_hello print_hello_bitrise print_hello_world", out)

	cmd = command.New(binPath(), "run", "print_hello", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello John Doe!")

	cmd = command.New(binPath(), "run", "print_hello_bitrise", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello Bitrise!")

	cmd = command.New(binPath(), "run", "print_hello_world", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello World!")
}

func Test_ModularConfig_Run_JSON_Logs(t *testing.T) {
	configPth := "modular_config_main.yml"

	cmd := command.New(binPath(), "run", "print_hello_world", "--config", configPth, "--output-format", "json")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello World!")
	checkRequiredStepBundle(t, out, "print")
}

func checkRequiredStepBundle(t *testing.T, log string, requiredStepBundle string) {
	lines := strings.Split(log, "\n")
	require.True(t, len(lines) > 0)

	var bitriseStartedEventLog struct {
		BitriseStartedEvent models.WorkflowRunPlan `json:"content"`
	}
	bitriseStartedLog := lines[0]
	require.NoError(t, json.Unmarshal([]byte(bitriseStartedLog), &bitriseStartedEventLog))
	bitriseStartedEvent := bitriseStartedEventLog.BitriseStartedEvent

	var usedStepBundles []string

	for _, workflowPlans := range bitriseStartedEvent.ExecutionPlan {
		for _, stepPlans := range workflowPlans.Steps {
			if stepPlans.StepBundleUUID != "" {
				stepBundlePlan := bitriseStartedEvent.StepBundlePlans[stepPlans.StepBundleUUID]
				usedStepBundles = append(usedStepBundles, stepBundlePlan.ID)
			}
		}
	}

	require.Equal(t, 1, len(usedStepBundles), bitriseStartedLog)
	require.EqualValues(t, requiredStepBundle, usedStepBundles[0], bitriseStartedLog)
}
