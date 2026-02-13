//go:build linux_only
// +build linux_only

package docker

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// Test_Docker_New_Syntax tests the new step-level containerization syntax
// where steps can directly reference execution_container and service_containers
// instead of using with-groups.
func Test_Docker_New_Syntax(t *testing.T) {
	testCases := map[string]struct {
		configPath          string
		workflowName        string
		requireErr          bool
		requireLogs         []string
		requiredLogPatterns []string
	}{
		"basic execution - step with execution container": {
			configPath:   "docker_new_syntax_basic_bitrise.yml",
			workflowName: "test-execution-container",
			requireErr:   false,
			requireLogs: []string{
				"Running in container",
				"Step is running in container:",
			},
		},
		"basic execution - step with service containers": {
			configPath:   "docker_new_syntax_basic_bitrise.yml",
			workflowName: "test-service-containers",
			requireErr:   false,
			requireLogs: []string{
				"Container (redis) is healthy",
			},
		},

		// Test 2: Container Lifecycle Management
		"lifecycle - container reuse across steps": {
			configPath:   "docker_new_syntax_lifecycle_bitrise.yml",
			workflowName: "test-lifecycle",
			requireErr:   false,
			requireLogs: []string{
				"Step 1 in container A",
				"Step 2 in container A",
				"Step 3 in container B",
				"Step 4 on host",
			},
		},

		// Test 3: Container Recreation Logic
		"recreation - recreate flag forces fresh container": {
			configPath:   "docker_new_syntax_recreate_bitrise.yml",
			workflowName: "test-recreate",
			requireErr:   false,
			requireLogs: []string{
				"Creating marker",
				"SUCCESS: Container reused - marker found",
				"SUCCESS: Container recreated - marker not found",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			cmd := command.New(
				testhelpers.BinPath(),
				"run",
				testCase.workflowName,
				"--config",
				testCase.configPath,
			)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if testCase.requireErr {
				require.Error(t, err, "Expected command to fail but it succeeded. Output:\n%s", out)
			} else {
				require.NoError(t, err, "Expected command to succeed but it failed. Output:\n%s", out)
			}

			for _, expectedLog := range testCase.requireLogs {
				require.Contains(t, out, expectedLog, "Expected log message not found in output:\n%s", out)
			}

			for _, expectedPattern := range testCase.requiredLogPatterns {
				require.Regexp(t, expectedPattern, out, "Expected log pattern not found in output:\n%s", out)
			}
		})
	}
}
