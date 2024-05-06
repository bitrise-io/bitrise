//go:build linux_only
// +build linux_only

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_Docker(t *testing.T) {
	testCases := map[string]struct {
		configPath    string
		inventoryPath string
		workflowName  string
		requireErr    bool
		requireLogs   []string
	}{
		"docker pull succeeds with existing image": {
			configPath:   "docker_pull_bitrise.yml",
			workflowName: "docker-pull-success",
			requireErr:   false,
			requireLogs:  []string{"Step is running in container:"},
		},
		"docker pull fails with non-existing image": {
			configPath:   "docker_pull_bitrise.yml",
			workflowName: "docker-pull-fails-404",
			requireErr:   true,
			requireLogs: []string{
				"Error during image pull",
				"Failed to pull image, retrying",
			},
		},
		"docker login fails when incorrect credentials are provided": {
			configPath:   "docker_pull_bitrise.yml",
			workflowName: "docker-login-fail",
			requireErr:   true,
			requireLogs: []string{
				"workflow has docker credentials provided, but the authentication failed",
			},
		},
		"docker login succeeds when correct credentials are provided": {
			configPath:    "docker_pull_bitrise.yml",
			inventoryPath: "docker_login_secrets.yml",
			workflowName:  "docker-login-success",
			requireErr:    false,
			requireLogs: []string{
				"Logging into docker registry:",
				"Step is running in container:",
				"--password [REDACTED]",
			},
		},
		"docker create fails when already-used port is provided": {
			configPath:   "docker_create_bitrise.yml",
			workflowName: "docker-create-fails-invalid-port",
			requireErr:   true,
			requireLogs: []string{
				"failed to start containers:",
				"bind: address already in use",
			},
		},
		"docker create succeeds when valid port is provided": {
			configPath:   "docker_create_bitrise.yml",
			workflowName: "docker-create-succeeds-valid-port",
			requireErr:   false,
			requireLogs: []string{
				"Step is running in container:",
			},
		},
		"docker create succeeds if false negative health check result is present": {
			configPath:   "docker_create_bitrise.yml",
			workflowName: "docker-create-succeeds-with-false-unhealthy-container",
			requireErr:   false,
			requireLogs: []string{
				"Container (bitrise-workflow-docker-create-succeeds-with-false-unhealthy-container) is unhealthy...",
				"Step is running in container: frolvlad/alpine-bash:latest",
			},
		},
		"docker create fails when invalid option is provided": {
			configPath:   "docker_create_bitrise.yml",
			workflowName: "docker-create-fails-invalid-option",
			requireErr:   true,
			requireLogs: []string{
				"unknown flag: --invalid-option",
				"Could not start the specified docker image for workflow:",
			},
		},
		"docker start fails with container throwing error": {
			configPath:   "docker_start_fails_bitrise.yml",
			workflowName: "docker-start-fails",
			requireErr:   true,
			requireLogs: []string{
				"nonexistent-command",
			},
		},
		"docker service start fails, build succeeds, service error is logged": {
			configPath:   "docker_service_bitrise.yml",
			workflowName: "docker-service-start-fails",
			requireErr:   false,
			requireLogs: []string{
				"Some services failed to start properly",
				"start docker container (failing-service): exit status 1",
				"nonexistent-command",
			},
		},
		"docker start services succeeds after retries": {
			configPath:   "docker_service_bitrise.yml",
			workflowName: "docker-service-start-succeeds-after-retries",
			requireErr:   false,
			requireLogs: []string{
				"Waiting for container (slow-bootin-service) to be healthy",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			cmd := command.New(binPath(), "run", testCase.workflowName, "--config", testCase.configPath)
			if testCase.inventoryPath != "" {
				cmd.GetCmd().Args = append(cmd.GetCmd().Args, "--inventory", testCase.inventoryPath)
			}

			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			if testCase.requireErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err, out)
			}
			for _, log := range testCase.requireLogs {
				require.Contains(t, out, log)
			}
		})
	}
}
