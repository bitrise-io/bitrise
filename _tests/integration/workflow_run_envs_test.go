//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_WorkflowRunEnvs(t *testing.T) {
	for _, tt := range []struct {
		name                string
		workflow            string
		usesSecrets         bool
		expectedToFail      bool
		expectedStepOutputs []string
	}{
		{
			name:           "Workflow run build status envs",
			workflow:       "workflow_run_envs_test",
			expectedToFail: true,
			expectedStepOutputs: []string{
				"BITRISE_BUILD_STATUS initially set to '0'\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"Failing skippable step\n",
				"Failing skippable step isn't not modifying BITRISE_BUILD_STATUS\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"Failing step\n",
				"BITRISE_BUILD_STATUS set to '1' on failure\nBITRISE_BUILD_STATUS: 1\nSTEPLIB_BUILD_STATUS: 1\n",
			},
		},
		{
			name:           "Step bundle build status envs",
			workflow:       "bundle_run_envs_test",
			expectedToFail: true,
			expectedStepOutputs: []string{
				"Failing step\nStep failure reason\n",
				"Build status failing inside bundle\nBITRISE_BUILD_STATUS: 1\nBITRISE_FAILED_STEP_TITLE: Failing step\nBITRISE_FAILED_STEP_ERROR_MESSAGE: Step failure reason\n",
				"Build status failing after bundle\nBITRISE_BUILD_STATUS: 1\nBITRISE_FAILED_STEP_TITLE: Failing step\nBITRISE_FAILED_STEP_ERROR_MESSAGE: Step failure reason\n",
			},
		},
		{
			name:           "Before, after workflow run build status envs",
			workflow:       "before_after_workflow_run_envs_test",
			expectedToFail: true,
			expectedStepOutputs: []string{
				"_before1 success step 1\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"_before1 failing skippable step\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"_before1 uccess step 2\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"_before2 success step\nBITRISE_BUILD_STATUS: 0\nSTEPLIB_BUILD_STATUS: 0\n",
				"Failing step\n",
				"_after1 failing step\nBITRISE_BUILD_STATUS: 1\nSTEPLIB_BUILD_STATUS: 1\n",
			},
		},
		{
			name:           "Build status envs test in run_if conditions",
			workflow:       "build_status_run_if_test",
			expectedToFail: true,
			expectedStepOutputs: []string{
				"Run if BITRISE_BUILD_STATUS is 0\n",
				"Run if not .IsBuildFailed\n",
				"Run if BITRISE_BUILD_STATUS is 1\n",
				"Run if .IsBuildFailed\n",
			},
		},
		{
			name:           "Failing step and failure reason envs test",
			workflow:       "failed_step_and_reason_envs_test",
			expectedToFail: true,
			expectedStepOutputs: []string{
				"Step failure reason\n",
				"BITRISE_FAILED_STEP_TITLE: Failing step\nBITRISE_FAILED_STEP_ERROR_MESSAGE: Step failure reason\n",
				"Run if BITRISE_FAILED_STEP_TITLE is 'Failing step'\n",
				"Run if BITRISE_FAILED_STEP_ERROR_MESSAGE is 'Step failure reason'\n",
			},
		},
		{
			name:           "Failing step and failure reason envs secret filtering test",
			workflow:       "failed_step_and_reason_envs_test",
			usesSecrets:    true,
			expectedToFail: true,
			expectedStepOutputs: []string{
				"Step [REDACTED]\n",
				"BITRISE_FAILED_STEP_TITLE: Failing step\nBITRISE_FAILED_STEP_ERROR_MESSAGE: Step [REDACTED]\n",
				"Run if BITRISE_FAILED_STEP_TITLE is 'Failing step'\n",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"run", "--output-format", "json", tt.workflow, "--config", "workflow_run_envs_test_bitrise.yml"}
			if tt.usesSecrets {
				args = append(args, "--inventory", "workflow_run_envs_test_secrets.yml")
			}
			cmd := command.New(binPath(), args...)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			stepOutputs := collectStepOutputs(out, t)

			if tt.expectedToFail {
				require.Error(t, err, out)
			} else {
				require.NoError(t, err, out)
			}

			if len(tt.expectedStepOutputs) > 0 {
				require.Equal(t, tt.expectedStepOutputs, stepOutputs)
			}
		})
	}
}
