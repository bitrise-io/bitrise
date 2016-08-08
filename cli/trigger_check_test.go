package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
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
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	t.Log("Default pattern defined &  Pull request mode")
	{
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", true)
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
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("No default pattern defined & Pull request mode")
	{
		workflowID, err := getWorkflowIDByPattern(config.TriggerMap, "master", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/a", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature/", true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "feature", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByPattern(config.TriggerMap, "test", true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}
}
