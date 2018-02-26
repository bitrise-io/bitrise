package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestSkipIfEmpty(t *testing.T) {
	t.Log("skip_if_empty=true && value=empty => should not add")
	{
		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  skip_if_empty:
    envs:
    - TEST: test
    - TEST:
      opts:
        skip_if_empty: true
    steps:
    - script:
        is_skippable: true
        title: "Envman add DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            if [ -z $TEST ] ; then
              echo "TEST shuld exist"
              exit 1
            fi
`
		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "skip_if_empty", config, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, 1, len(buildRunResults.SuccessSteps))
		require.Equal(t, 0, len(buildRunResults.FailedSteps))
		require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
		require.Equal(t, 0, len(buildRunResults.SkippedSteps))
	}

	t.Log("skip_if_empty=false && value=empty => should add")
	{
		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  skip_if_empty:
    envs:
    - TEST: test
    - TEST:
      opts:
        skip_if_empty: false
    steps:
    - script:
        is_skippable: true
        title: "Envman add DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            if [ ! -z $TEST ] ; then
              echo "TEST env shuld not exist"
              exit 1
            fi
`
		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "skip_if_empty", config, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, 1, len(buildRunResults.SuccessSteps))
		require.Equal(t, 0, len(buildRunResults.FailedSteps))
		require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
		require.Equal(t, 0, len(buildRunResults.SkippedSteps))
	}
}

func TestDeleteEnvironment(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    steps:
    - script:
        is_skippable: true
        title: "Envman add DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            envman add --key DELETE_TEST --value "delete test"
    - script:
        title: "Test env DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "DELETE_TEST: $DELETE_TEST"
            if [ -z "$DELETE_TEST" ] ; then
              exit 1
            fi
    - script:
        title: "Delete env DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            envman add --key DELETE_TEST --value ""
    - script:
        title: "Test env DELETE_TEST"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "DELETE_TEST: $DELETE_TEST"
            if [ ! -z "$DELETE_TEST" ] ; then
              exit 1
            fi
`
	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 4, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestStepOutputsInTemplate(t *testing.T) {
	inventoryStr := `
envs:
- TEMPLATE_TEST0: "true"
`
	inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.NoError(t, err)

	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - TEMPLATE_TEST1: "true"

workflows:
  test:
    envs:
    - TEMPLATE_TEST2: "true"
    steps:
    - script:
        title: "Envman add"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            envman add --key TEMPLATE_TEST3 --value "true"
    - script:
        title: "TEMPLATE_TEST0"
        run_if: |-
          {{enveq "TEMPLATE_TEST0" "true"}}
    - script:
        title: "TEMPLATE_TEST1"
        run_if: |-
          {{enveq "TEMPLATE_TEST1" "true"}}
    - script:
        title: "TEMPLATE_TEST2"
        run_if: |-
          {{enveq "TEMPLATE_TEST2" "true"}}
    - script:
        title: "TEMPLATE_TEST3"
        run_if: |-
          {{enveq "TEMPLATE_TEST3" "true"}}
    - script:
        title: "TEMPLATE_TEST_NO_VALUE"
        run_if: |-
          {{enveq "TEMPLATE_TEST_NO_VALUE" "true"}}
`

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.NoError(t, err)
	require.Equal(t, 5, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))
}

func TestFailedStepOutputs(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    steps:
    - script:
        is_skippable: true
        title: "Envman add"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            envman add --key FAILED_OUTPUT_TEST --value "failed step output"
            exit 1
    - script:
        title: "Test failed output"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "FAILED_OUTPUT_TEST: $FAILED_OUTPUT_TEST"
            if [[ "$FAILED_OUTPUT_TEST" != "failed step output" ]] ; then
              exit 1
            fi
`
	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestBitriseSourceDir(t *testing.T) {
	currPth, err := pathutil.NormalizedOSTempDirPath("bitrise_source_dir_test")
	require.NoError(t, err)

	testPths := []string{}
	for i := 0; i < 4; i++ {
		testPth := filepath.Join(currPth, fmt.Sprintf("_test%d", i))
		require.NoError(t, os.RemoveAll(testPth))
		require.NoError(t, os.Mkdir(testPth, 0777))

		// eval symlinks: the Go generated temp folder on OS X is a symlink
		//  from /var/ to /private/var/
		testPth, err = filepath.EvalSymlinks(testPth)
		require.NoError(t, err)

		defer func() { require.NoError(t, os.RemoveAll(testPth)) }()

		testPths = append(testPths, testPth)
	}

	t.Log("BITRISE_SOURCE_DIR defined in Secret")
	{
		inventoryStr := `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "BITRISE_SOURCE_DIR: $BITRISE_SOURCE_DIR"
            if [[ "$BITRISE_SOURCE_DIR" != "` + testPths[0] + `" ]] ; then
              exit 1
            fi
`
		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("BITRISE_SOURCE_DIR defined in Secret, and in App")
	{
		inventoryStr := `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - BITRISE_SOURCE_DIR: "` + testPths[1] + `"

workflows:
  test:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "BITRISE_SOURCE_DIR: $BITRISE_SOURCE_DIR"
            if [[ "$BITRISE_SOURCE_DIR" != "` + testPths[1] + `" ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("BITRISE_SOURCE_DIR defined in Secret, App and Workflow")
	{
		inventoryStr := `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - BITRISE_SOURCE_DIR: "` + testPths[1] + `"

workflows:
  test:
    envs:
    - BITRISE_SOURCE_DIR: "` + testPths[2] + `"
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "BITRISE_SOURCE_DIR: $BITRISE_SOURCE_DIR"
            if [[ "$BITRISE_SOURCE_DIR" != "` + testPths[2] + `" ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("BITRISE_SOURCE_DIR defined in secret, App, Workflow and Step")
	{
		inventoryStr := `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - BITRISE_SOURCE_DIR: "` + testPths[1] + `"

workflows:
  test:
    envs:
    - BITRISE_SOURCE_DIR: "` + testPths[2] + `"
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            envman add --key BITRISE_SOURCE_DIR --value ` + testPths[3] + `
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "BITRISE_SOURCE_DIR: $BITRISE_SOURCE_DIR"
            if [[ "$BITRISE_SOURCE_DIR" != "` + testPths[3] + `" ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}
}

func TestEnvOrders(t *testing.T) {
	t.Log("Only secret env - secret env should be use")
	{
		inventoryStr := `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "ENV_ORDER_TEST: $ENV_ORDER_TEST"
            if [[ "$ENV_ORDER_TEST" != "should be the 1." ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("Secret env & app env also defined - app env should be use")
	{
		inventoryStr := `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - ENV_ORDER_TEST: "should be the 2."

workflows:
  test:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "ENV_ORDER_TEST: $ENV_ORDER_TEST"
            if [[ "$ENV_ORDER_TEST" != "should be the 2." ]] ; then
              exit 1
            fi

`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("Secret env & app env && workflow env also defined - workflow env should be use")
	{
		inventoryStr := `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - ENV_ORDER_TEST: "should be the 2."

workflows:
  test:
    envs:
    - ENV_ORDER_TEST: "should be the 3."
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "ENV_ORDER_TEST: $ENV_ORDER_TEST"
            if [[ "$ENV_ORDER_TEST" != "should be the 3." ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}

	t.Log("Secret env & app env && workflow env && step output env also defined - step output env should be use")
	{
		inventoryStr := `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - ENV_ORDER_TEST: "should be the 2."

workflows:
  test:
    envs:
    - ENV_ORDER_TEST: "should be the 3."
    steps:
    - script:
        inputs:
        - content: envman add --key ENV_ORDER_TEST --value "should be the 4."
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "ENV_ORDER_TEST: $ENV_ORDER_TEST"
            if [[ "$ENV_ORDER_TEST" != "should be the 4." ]] ; then
              exit 1
            fi
`

		require.NoError(t, configs.InitPaths())

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
		require.NoError(t, err)
	}
}

// Test - Bitrise activateAndRunWorkflow
// If workflow contains no steps
func Test0Steps1Workflows(t *testing.T) {
	workflow := models.WorkflowModel{}

	require.NoError(t, os.Setenv("BITRISE_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("BITRISE_BUILD_STATUS")) }()

	require.NoError(t, os.Setenv("STEPLIB_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("STEPLIB_BUILD_STATUS")) }()

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"zero_steps": workflow,
		},
	}

	_, err := config.Validate()
	require.NoError(t, err)

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}

	require.NoError(t, configs.InitPaths())

	buildRunResults, err = runWorkflowWithConfiguration(time.Now(), "zero_steps", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 0, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise activateAndRunWorkflow
// Workflow contains before and after workflow, and no one contains steps
func Test0Steps3WorkflowsBeforeAfter(t *testing.T) {
	require.NoError(t, os.Setenv("BITRISE_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("BITRISE_BUILD_STATUS")) }()

	require.NoError(t, os.Setenv("STEPLIB_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("STEPLIB_BUILD_STATUS")) }()

	beforeWorkflow := models.WorkflowModel{}
	afterWorkflow := models.WorkflowModel{}

	workflow := models.WorkflowModel{
		BeforeRun: []string{"before"},
		AfterRun:  []string{"after"},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target": workflow,
			"before": beforeWorkflow,
			"after":  afterWorkflow,
		},
	}

	_, err := config.Validate()
	require.NoError(t, err)

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}

	buildRunResults, err = activateAndRunWorkflow(
		"target", workflow, config, buildRunResults,
		&[]envmanModels.EnvironmentItemModel{}, []envmanModels.EnvironmentItemModel{},
		"",
	)
	require.NoError(t, err)
	require.Equal(t, 0, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise Validate workflow
// Workflow contains before and after workflow, and no one contains steps, but circular wofklow dependecy exist, which should fail
func Test0Steps3WorkflowsCircularDependency(t *testing.T) {
	require.NoError(t, os.Setenv("BITRISE_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("BITRISE_BUILD_STATUS")) }()

	require.NoError(t, os.Setenv("STEPLIB_BUILD_STATUS", "0"))
	defer func() { require.NoError(t, os.Unsetenv("STEPLIB_BUILD_STATUS")) }()

	beforeWorkflow := models.WorkflowModel{
		BeforeRun: []string{"target"},
	}

	afterWorkflow := models.WorkflowModel{}

	workflow := models.WorkflowModel{
		BeforeRun: []string{"before"},
		AfterRun:  []string{"after"},
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"target": workflow,
			"before": beforeWorkflow,
			"after":  afterWorkflow,
		},
	}

	_, err := config.Validate()
	require.Error(t, err)

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise activateAndRunWorkflow
// Trivial test with 1 workflow
func Test1Workflows(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  trivial_fail:
    steps:
    - script:
        title: Should success
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2
    - script:
        title: Should success
    - script:
        title: Should fail
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2
    - script:
        title: Should success
        is_always_run: true
    - script:
        title: Should skipped
  `
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	workflow, found := config.Workflows["trivial_fail"]
	require.Equal(t, true, found)

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}

	buildRunResults, err = activateAndRunWorkflow(
		"trivial_fail", workflow, config, buildRunResults,
		&[]envmanModels.EnvironmentItemModel{}, []envmanModels.EnvironmentItemModel{},
		"",
	)
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise activateAndRunWorkflow
// Trivial test with before, after workflows
func Test3Workflows(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before1:
    steps:
    - script:
        title: Should success
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2
    - script:
        title: Should success

  before2:
    steps:
    - script:
        title: Should success

  target:
    before_run:
    - before1
    - before2
    after_run:
    - after1
    - after2
    steps:
    - script:
        title: Should fail
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2

  after1:
    steps:
    - script:
        title: Should fail
        is_always_run: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2

  after2:
    steps:
    - script:
        title: Should be skipped
  `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 2, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise ConfigModelFromYAMLBytes
// Workflow contains before and after workflow, and no one contains steps, but circular wofklow dependecy exist, which should fail
func TestRefeneceCycle(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before1:
    before_run:
    - before2

  before2:
    before_run:
    - before1

  target:
    before_run:
    - before1
    - before2
  `
	_, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Error(t, err)
	require.Equal(t, 0, len(warnings))
}

// Test - Bitrise BuildStatusEnv
// Checks if BuildStatusEnv is set correctly
func TestBuildStatusEnv(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before1:
    steps:
    - script:
        title: Should success
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2
    - script:
        title: Should success

  before2:
    steps:
    - script:
        title: Should success

  target:
    steps:
    - script:
        title: Should success
        inputs:
        - content: |
            #!/bin/bash
            set -v
            if [[ "$BITRISE_BUILD_STATUS" != "0" ]] ; then
              exit 1
            fi
            if [[ "$STEPLIB_BUILD_STATUS" != "0" ]] ; then
              exit 1
            fi
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo 'This is a before workflow'
            exit 2
    - script:
        title: Should success
        inputs:
        - content: |
            #!/bin/bash
            set -v
            if [[ "$BITRISE_BUILD_STATUS" != "0" ]] ; then
              exit 1
            fi
            if [[ "$STEPLIB_BUILD_STATUS" != "0" ]] ; then
              exit 1
            fi
    - script:
        title: Should fail
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 1
    - script:
        title: Should success
        is_always_run: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            if [[ "$BITRISE_BUILD_STATUS" != "1" ]] ; then
              echo "should fail"
            fi
            if [[ "$STEPLIB_BUILD_STATUS" != "1" ]] ; then
              echo "should fail"
            fi
    - script:
        title: Should skipped
  `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise activateAndRunWorkflow
// Trivial fail test
func TestFail(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    steps:
    - script:
        title: Should success
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2
    - script:
        title: Should success
    - script:
        title: Should fail
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 1
    - script:
        title: Should skipped
    - script:
        title: Should success
        is_always_run: true
    `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise activateAndRunWorkflow
// Trivial success test
func TestSuccess(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    steps:
    - script:
        title: Should success
    `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise BuildStatusEnv
// Checks if BuildStatusEnv is set correctly
func TestBuildFailedMode(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before1:
    title: before1
    steps:
    - script:
        title: Should success
    - script:
        title: Should fail
        inputs:
        - content: |
            #!/bin/bash
            set -v
            exit 2

  before2:
    title: before2
    steps:
    - script:
        title: Should skipped

  target:
    title: target
    before_run:
    - before1
    - before2
    steps:
    - script:
        title: Should skipped
    `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 2, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise Environments
// Trivial test for workflow environment handling
// Before workflows env should be visible in target and after workflow
func TestWorkflowEnvironments(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:
    envs:
    - BEFORE_ENV: beforeenv

  target:
    title: target
    before_run:
    - before
    after_run:
    - after
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            if [[ "$BEFORE_ENV" != "beforeenv" ]] ; then
              exit 1
            fi

  after:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            if [[ "$BEFORE_ENV" != "beforeenv" ]] ; then
              exit 1
            fi
    `

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise Environments
// Test for same env in before and target workflow, actual workflow should overwrite environemnt and use own value
func TestWorkflowEnvironmentOverWrite(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:
    envs:
    - ENV: env1
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo ${ENV}
            if [[ "$ENV" != "env1" ]] ; then
              exit 1
            fi

  target:
    title: target
    envs:
    - ENV: env2
    before_run:
    - before
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo ${ENV}
            if [[ "$ENV" != "env2" ]] ; then
              exit 1
            fi
`

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise Environments
// Target workflows env should be visible in before and after workflow
func TestTargetDefinedWorkflowEnvironment(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo ${ENV}
            if [[ "$ENV" != "targetenv" ]] ; then
              exit 3
            fi

  target:
    title: target
    envs:
    - ENV: targetenv
    before_run:
    - before
`

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Test - Bitrise Environments
// Step input should visible only for actual step and invisible for other steps
func TestStepInputEnvironment(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:
    steps:
    - script@1.1.3:
        inputs:
        - working_dir: $HOME

  target:
    title: target
    before_run:
    - before
    steps:
    - script@1.1.3:
        title: "${working_dir} should not exist"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            env
            if [ ! -z "$working_dir" ] ; then
              echo ${working_dir}
              exit 3
            fi
`

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	if os.Getenv("working_dir") != "" {
		require.Equal(t, nil, os.Unsetenv("working_dir"))
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Outputs exported with `envman add` should be accessible for subsequent Steps.
func TestStepOutputEnvironment(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  out-test:
    title: Output Test
    steps:
    - script:
        inputs:
        - content: envman -l=debug add --key MY_TEST_1 --value 'Test value 1'
    - script:
        inputs:
        - content: |-
            if [[ "${MY_TEST_1}" != "Test value 1" ]] ; then
              echo " [!] MY_TEST_1 invalid: ${MY_TEST_1}"
              exit 1
            fi
    - script:
        title: Should fail
        inputs:
        - content: |-
            envman add --key MY_TEST_2 --value 'Test value 2'
            # exported output, but test fails
            exit 22
    - script:
        is_always_run: true
        inputs:
        - content: |-
            if [[ "${MY_TEST_2}" != "Test value 2" ]] ; then
              exit 1
            fi
`

	require.NoError(t, configs.InitPaths())

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["out-test"]
	require.Equal(t, true, found)

	_, err = config.Validate()
	require.NoError(t, err)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "out-test", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))

	// the exported output envs should NOT be exposed here, should NOT be available!
	require.Equal(t, "", os.Getenv("MY_TEST_1"))
	require.Equal(t, "", os.Getenv("MY_TEST_2"))

	// standard, Build Status ENV test
	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

func TestLastWorkflowIDInConfig(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:

  target:
    title: target
    before_run:
    - before
    after_run:
    - after1

  after1:
    after_run:
    - after2

  after2:
  `
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	last, err := lastWorkflowIDInConfig("target", config)
	require.NoError(t, err)
	require.Equal(t, "after2", last)
}
