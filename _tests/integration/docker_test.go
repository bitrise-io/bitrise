//go:build linux_only
// +build linux_only

package integration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/ryanuber/go-glob"
	"github.com/stretchr/testify/require"
)

func Test_Docker(t *testing.T) {
	testCases := map[string]struct {
		configPath          string
		inventoryPath       string
		workflowName        string
		requireErr          bool
		requireLogs         []string
		requiredLogPatterns []string
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
				"docker credentials provided, but the authentication failed",
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
				"Step is running in container: frolvlad/alpine-bash:latest",
			},
			requiredLogPatterns: []string{
				"*Container (bitrise-workflow-*) is unhealthy...*",
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
				"Waiting for container (slow-booting-service) to be healthy",
			},
		},
		"docker start container and services with credentials": {
			configPath:    "docker_multiple_containers_bitrise.yml",
			workflowName:  "docker-login-multiple-containers",
			inventoryPath: "docker_multiple_containers_secrets.yml",
			requireErr:    false,
			requireLogs: []string{
				"Container (service_1_container) is healthy...",
				"Container (service_2_container) is healthy...",
				"Step is running in container: localhost:5001/healthy-image",
			},
		},
		"docker stops containers when needed": {
			configPath:   "docker_stop_containers_bitrise.yml",
			workflowName: "docker-stops-containers",
			requireErr:   false,
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
			for _, logPattern := range testCase.requiredLogPatterns {
				contains := glob.Glob(logPattern, out)
				require.True(t, contains, out)
			}
		})
	}
}

func Test_Docker_JSON_Logs(t *testing.T) {
	testCases := map[string]struct {
		workflowName           string
		configPath             string
		inventoryPath          string
		requiredContainerImage string
		requiredServiceImages  []string
	}{
		"With group with step execution and service containers": {
			workflowName:           "docker-login-multiple-containers",
			configPath:             "docker_multiple_containers_bitrise.yml",
			inventoryPath:          "docker_multiple_containers_secrets.yml",
			requiredContainerImage: "localhost:5001/healthy-image",
			requiredServiceImages: []string{
				"localhost:5002/healthy-image",
				"localhost:5003/healthy-image",
			},
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			cmd := command.New(binPath(), "run", testCase.workflowName, "--config", testCase.configPath, "--inventory", testCase.inventoryPath, "--output-format", "json")
			out, _ := cmd.RunAndReturnTrimmedCombinedOutput()
			//require.NoError(t, err, out)
			checkRequiredContainers(t, out, testCase.requiredContainerImage, testCase.requiredServiceImages)
		})
	}
}

func checkRequiredContainers(t *testing.T, log string, requiredContainerImage string, requiredServiceImages []string) {
	lines := strings.Split(log, "\n")
	require.True(t, len(lines) > 0)

	var bitriseStartedEvent models.WorkflowRunPlan
	for _, line := range lines {
		var eventLogStruct struct {
			EventType string                 `json:"event_type"`
			Content   models.WorkflowRunPlan `json:"content"`
		}
		require.NoError(t, json.Unmarshal([]byte(line), &eventLogStruct))
		if eventLogStruct.EventType == "bitrise_started" {
			bitriseStartedEvent = eventLogStruct.Content
			break
		}
	}

	var usedContainerImages []string
	var usedServiceImages []string

	for _, workflowPlans := range bitriseStartedEvent.ExecutionPlan {
		for _, stepPlans := range workflowPlans.Steps {
			if stepPlans.WithGroupUUID != "" {
				withGroupPlan := bitriseStartedEvent.WithGroupPlans[stepPlans.WithGroupUUID]

				usedContainerImages = append(usedContainerImages, withGroupPlan.Container.Image)
				for _, servicePlan := range withGroupPlan.Services {
					usedServiceImages = append(usedServiceImages, servicePlan.Image)
				}
			}
		}
	}

	require.Equal(t, 1, len(usedContainerImages), log)
	require.EqualValues(t, requiredContainerImage, usedContainerImages[0], log)
	require.EqualValues(t, requiredServiceImages, usedServiceImages, log)
}
