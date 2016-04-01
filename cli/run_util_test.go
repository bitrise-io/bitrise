package cli

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestExpandEnvs(t *testing.T) {
	configStr := `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    envs:
    - ENV0: "Hello"
    - ENV1: "$ENV0 world"
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            envman add --key ENV2 --value "$ENV1 !"
    - script:
        inputs:
        - content: |
            #!/bin/bash
            echo "ENV2: $ENV2"
            if [ "$ENV2" != "Hello world !" ] ; then
              echo "Actual ($ENV2), excpected (Hello world !)"
              exit 1
            fi
`
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestEvaluateInputs(t *testing.T) {
	configStr := `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    envs:
    - TEST_KEY: "test value"
    steps:
    - script:
        title: "Template test"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            {{if .IsCI}}
            exit 1
            {{else}}
            exit 0
            {{end}}
          opts:
            is_template: true
    - script:
        title: "Template test"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            {{if enveq "TEST_KEY" "test value"}}
            exit 0
            {{else}}
            exit 1
            {{end}}
          opts:
            is_template: true
`
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, nil, err)
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
}

func TestGetWorkflowIDByPattern(t *testing.T) {
	configStr := `
trigger_map:
- pattern: master
  is_pull_request_allowed: false
  workflow: master
- pattern: feature/*
  is_pull_request_allowed: true
  workflow: feature
- pattern: "*"
  is_pull_request_allowed: true
  workflow: primary

workflows:
  test:
  master:
  feature:
  primary:
`
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	t.Log("Default pattern defined & Non pull request mode")
	{
		configs.IsPullRequestMode = false

		workflowID, err := GetWorkflowIDByPattern(config, "master")
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/a")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature")
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "test")
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	t.Log("Default pattern defined &  Pull request mode")
	{
		configs.IsPullRequestMode = true

		workflowID, err := GetWorkflowIDByPattern(config, "master")
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/a")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature")
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "test")
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	configStr = `
  trigger_map:
  - pattern: master
    is_pull_request_allowed: false
    workflow: master
  - pattern: feature/*
    is_pull_request_allowed: true
    workflow: feature

  workflows:
    test:
    master:
    feature:
    primary:
  `
	config, warnings, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	t.Log("No default pattern defined & Non pull request mode")
	{
		configs.IsPullRequestMode = false

		workflowID, err := GetWorkflowIDByPattern(config, "master")
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/a")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature")
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "test")
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("No default pattern defined & Pull request mode")
	{
		configs.IsPullRequestMode = true

		workflowID, err := GetWorkflowIDByPattern(config, "master")
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/a")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature/")
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "feature")
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = GetWorkflowIDByPattern(config, "test")
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}
}

func TestGetBitriseConfigFromBase64Data(t *testing.T) {
	configStr := `
format_version: 0.9.10
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    title: target
`
	configBytes := []byte(configStr)
	configBase64Str := base64.StdEncoding.EncodeToString(configBytes)

	config, warnings, err := GetBitriseConfigFromBase64Data(configBase64Str)
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	require.Equal(t, "0.9.10", config.FormatVersion)
	require.Equal(t, "https://github.com/bitrise-io/bitrise-steplib.git", config.DefaultStepLibSource)

	workflow, found := config.Workflows["target"]
	require.Equal(t, true, found)
	require.Equal(t, "target", workflow.Title)
}

func TestGetInventoryFromBase64Data(t *testing.T) {
	inventoryStr := `
envs:
  - MY_HOME: $HOME
    opts:
      is_expand: true
`
	inventoryBytes := []byte(inventoryStr)
	inventoryBase64Str := base64.StdEncoding.EncodeToString(inventoryBytes)

	inventory, err := GetInventoryFromBase64Data(inventoryBase64Str)
	require.NoError(t, err)

	env := inventory[0]

	key, value, err := env.GetKeyValuePair()
	require.NoError(t, err)
	require.Equal(t, "MY_HOME", key)
	require.Equal(t, "$HOME", value)

	opts, err := env.GetOptions()
	require.NoError(t, err)
	require.Equal(t, true, *opts.IsExpand)
}

func TestInvalidStepID(t *testing.T) {
	configStr := `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    title: Invalid step id
    steps:
    - invalid-step:
    - invalid-step:
    - invalid-step:
`
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	results, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, 1, len(results.StepmanUpdates))
}
