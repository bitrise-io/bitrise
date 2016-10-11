package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func TestMigratePatternToParams(t *testing.T) {
	t.Log("converts pattern in NON PR MODE to push-branch param")
	{
		isPullRequestMode := false
		params := RunAndTriggerParamsModel{
			TriggerPattern: "master",
		}

		convertedParams := migratePatternToParams(params, isPullRequestMode)

		require.Equal(t, "master", convertedParams.PushBranch)
		require.Equal(t, "", convertedParams.PRSourceBranch)
		require.Equal(t, "", convertedParams.PRTargetBranch)
		require.Equal(t, "", convertedParams.TriggerPattern)
		require.Equal(t, "", convertedParams.Tag)

		require.Equal(t, "", convertedParams.WorkflowToRunID)
		require.Equal(t, "", convertedParams.Format)
		require.Equal(t, "", convertedParams.BitriseConfigPath)
		require.Equal(t, "", convertedParams.BitriseConfigBase64Data)
		require.Equal(t, "", convertedParams.InventoryPath)
		require.Equal(t, "", convertedParams.InventoryBase64Data)
	}

	t.Log("converts pattern in PR MODE to pr-source-branch param")
	{
		isPullRequestMode := true
		params := RunAndTriggerParamsModel{
			TriggerPattern: "master",
		}

		convertedParams := migratePatternToParams(params, isPullRequestMode)

		require.Equal(t, "", convertedParams.PushBranch)
		require.Equal(t, "master", convertedParams.PRSourceBranch)
		require.Equal(t, "", convertedParams.PRTargetBranch)
		require.Equal(t, "", convertedParams.TriggerPattern)
		require.Equal(t, "", convertedParams.Tag)

		require.Equal(t, "", convertedParams.WorkflowToRunID)
		require.Equal(t, "", convertedParams.Format)
		require.Equal(t, "", convertedParams.BitriseConfigPath)
		require.Equal(t, "", convertedParams.BitriseConfigBase64Data)
		require.Equal(t, "", convertedParams.InventoryPath)
		require.Equal(t, "", convertedParams.InventoryBase64Data)
	}

	t.Log("only modifies PushBranch, PRSourceBranch, PRTargetBranch, TriggerPattern")
	{
		isPullRequestMode := true
		params := RunAndTriggerParamsModel{
			PushBranch:     "feature/login",
			PRSourceBranch: "feature/landing",
			PRTargetBranch: "develop",
			Tag:            "0.9.0",
			TriggerPattern: "master",

			WorkflowToRunID:         "primary",
			Format:                  "json",
			BitriseConfigPath:       "bitrise.yml",
			BitriseConfigBase64Data: "base64-bitrise.yml",
			InventoryPath:           "inventory.yml",
			InventoryBase64Data:     "base64-inventory.yml",
		}

		convertedParams := migratePatternToParams(params, isPullRequestMode)

		require.Equal(t, "", convertedParams.PushBranch)
		require.Equal(t, "master", convertedParams.PRSourceBranch)
		require.Equal(t, "", convertedParams.PRTargetBranch)
		require.Equal(t, "", convertedParams.TriggerPattern)
		require.Equal(t, "", convertedParams.Tag)

		require.Equal(t, "primary", convertedParams.WorkflowToRunID)
		require.Equal(t, "json", convertedParams.Format)
		require.Equal(t, "bitrise.yml", convertedParams.BitriseConfigPath)
		require.Equal(t, "base64-bitrise.yml", convertedParams.BitriseConfigBase64Data)
		require.Equal(t, "inventory.yml", convertedParams.InventoryPath)
		require.Equal(t, "base64-inventory.yml", convertedParams.InventoryBase64Data)
	}
}

func TestGetWorkflowIDByParamsInCompatibleMode_new_param_test(t *testing.T) {
	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- push_branch: master
  workflow: master

workflows:
  master:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{PushBranch: "master"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- pull_request_source_branch: feature/*
  workflow: test

workflows:
  test:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			PRSourceBranch: "feature/login",
			PRTargetBranch: "develop",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "test", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- pull_request_target_branch: deploy_*
  workflow: release

workflows:
  release:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			PRSourceBranch: "master",
			PRTargetBranch: "deploy_1_0_0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "release", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- pull_request_source_branch: feature/*
  pull_request_target_branch: develop
  workflow: test

workflows:
  test:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			PRSourceBranch: "feature/login",
			PRTargetBranch: "develop",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "test", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- tag: 1.*
  workflow: deploy

workflows:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "1.0.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "deploy", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- tag: *.*.*
  workflow: deploy

workflows:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "1.0.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "deploy", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- tag: v*.*.*
  workflow: deploy

workflows:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "v1.0.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "deploy", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- tag: *.*.*
  workflow: deploy

workflows:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "1.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Error(t, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("it works with new params")
	{
		configStr := `
trigger_map:
- tag: v*.*.*
  workflow: deploy

workflows:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "v1.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Error(t, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("it works with complex trigger map")
	{
		configStr := `
trigger_map:
- pattern: feature/*
  workflow: test
- push_branch: feature/*
  workflow: test
- pull_request_source_branch: feature/*
  pull_request_target_branch: develop
  workflow: test
- tag: 1.*
  workflow: deploy

workflows:
  test:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			PRSourceBranch: "feature/login",
			PRTargetBranch: "develop",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "test", workflowID)
	}

	t.Log("it works with complex trigger map")
	{
		configStr := `
trigger_map:
- pattern: feature/*
  workflow: test
- push_branch: feature/*
  workflow: test
- pull_request_source_branch: feature/*
  pull_request_target_branch: develop
  workflow: test
- tag: 1.*
  workflow: deploy

workflows:
  test:
  deploy:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		params := RunAndTriggerParamsModel{
			Tag: "1.0.0",
		}
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, params, false)
		require.Equal(t, nil, err)
		require.Equal(t, "deploy", workflowID)
	}
}

func TestGetWorkflowIDByParamsInCompatibleMode_migration_test(t *testing.T) {
	t.Log("deprecated code push trigger item")
	{
		configStr := `
trigger_map:
- pattern: master
  is_pull_request_allowed: false
  workflow: master

workflows:
  master:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		t.Log("it works with deprecated pattern")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}

		t.Log("it works with new params")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{PushBranch: "master"}, false)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}

		t.Log("it works with new params")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{PushBranch: "master"}, true)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}
	}

	t.Log("deprecated pr trigger item")
	{
		configStr := `
trigger_map:
- pattern: master
  is_pull_request_allowed: true
  workflow: master

workflows:
  master:
`

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		t.Log("it works with deprecated pattern")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}

		t.Log("it works with new params")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{PushBranch: "master"}, false)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}

		t.Log("it works with new params")
		{
			workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{PushBranch: "master"}, true)
			require.Equal(t, nil, err)
			require.Equal(t, "master", workflowID)
		}
	}
}

func TestGetWorkflowIDByParamsInCompatibleMode_old_tests(t *testing.T) {
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
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)
	}

	t.Log("Default pattern defined &  Pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "primary", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, true)
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
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "master", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, false)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, false)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}

	t.Log("No default pattern defined & Pull request mode")
	{
		workflowID, err := getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "master"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/a"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature/"}, true)
		require.Equal(t, nil, err)
		require.Equal(t, "feature", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "feature"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)

		workflowID, err = getWorkflowIDByParamsInCompatibleMode(config.TriggerMap, RunAndTriggerParamsModel{TriggerPattern: "test"}, true)
		require.NotEqual(t, nil, err)
		require.Equal(t, "", workflowID)
	}
}
