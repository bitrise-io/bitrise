package cli

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

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
  is_pull_request_allowed: false
  workflow: primary

workflows:
  test:
  master:
  feature:
  primary:
`

	//
	// Default pattern defined
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	// Non pull request mode
	IsPullRequestMode = false

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

	// Pull request mode
	IsPullRequestMode = true

	workflowID, err = GetWorkflowIDByPattern(config, "master")
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

	//
	// No default pattern defined
	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	// Non pull request mode
	IsPullRequestMode = false

	workflowID, err = GetWorkflowIDByPattern(config, "master")
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
	require.Equal(t, "feature", workflowID)

	workflowID, err = GetWorkflowIDByPattern(config, "test")
	require.Equal(t, nil, err)
	require.Equal(t, "test", workflowID)

	// Pull request mode
	IsPullRequestMode = true

	workflowID, err = GetWorkflowIDByPattern(config, "master")
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
	t.Log("Config:", configBase64Str)

	config, err := GetBitriseConfigFromBase64Data(configBase64Str)
	if err != nil {
		t.Fatal("Failed to get config from base 64 data, err:", err)
	}

	if config.FormatVersion != "0.9.10" {
		t.Fatal("Invalid FormatVersion")
	}
	if config.DefaultStepLibSource != "https://github.com/bitrise-io/bitrise-steplib.git" {
		t.Fatal("Invalid FormatVersion")
	}

	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("Failed to found workflow")
	}
	if workflow.Title != "target" {
		t.Fatal("Invalid workflow.Title")
	}
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
	t.Log("Inventory:", inventoryBase64Str)

	inventory, err := GetInventoryFromBase64Data(inventoryBase64Str)
	if err != nil {
		t.Fatal("Failed to get inventory from base 64 data, err:", err)
	}

	env := inventory[0]

	key, value, err := env.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get env key-value pair, err:", err)
	}

	if key != "MY_HOME" {
		t.Fatal("Invalid key")
	}
	if value != "$HOME" {
		t.Fatal("Invalid value")
	}

	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal("Failed to get env options, err:", err)
	}

	if *opts.IsExpand != true {
		t.Fatal("Invalid IsExpand")
	}
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

	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}

	results, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, 1, len(results.StepmanUpdates))
}
