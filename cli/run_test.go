package cli

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"

	cliAnalytics "github.com/bitrise-io/bitrise/analytics"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/stretchr/testify/assert"
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "skip_if_empty"}
		runner := NewWorkflowRunner(runConfig)
		buildRunResults, err := runner.runWorkflows(noOpTracker{})
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
		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "skip_if_empty"}
		runner := NewWorkflowRunner(runConfig)
		buildRunResults, err := runner.runWorkflows(noOpTracker{})
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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "test"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "test"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestBitriseSourceDir(t *testing.T) {
	currPth, err := pathutil.NormalizedOSTempDirPath("bitrise_source_dir_test")
	require.NoError(t, err)

	var testPths []string
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
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

		config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
		require.NoError(t, err)
		require.Equal(t, 0, len(warnings))

		require.NoError(t, configs.InitPaths())

		runConfig := RunConfig{Config: config, Workflow: "test", Secrets: inventory.Envs}
		runner := NewWorkflowRunner(runConfig)
		_, err = runner.runWorkflows(noOpTracker{})
		require.NoError(t, err)
	}
}

// If workflow contains no steps
func Test0Steps1Workflows(t *testing.T) {
	workflow := models.WorkflowModel{}

	t.Setenv("BITRISE_BUILD_STATUS", "0")
	t.Setenv("STEPLIB_BUILD_STATUS", "0")

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

	runConfig := RunConfig{Config: config, Workflow: "zero_steps"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err = runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 0, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Workflow contains before and after workflow, and no one contains steps
func Test0Steps3WorkflowsBeforeAfter(t *testing.T) {
	t.Setenv("BITRISE_BUILD_STATUS", "0")
	t.Setenv("STEPLIB_BUILD_STATUS", "0")

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

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err = runner.runWorkflows(noOpTracker{})

	require.NoError(t, err)
	require.Equal(t, 0, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "trivial_fail"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err = runner.runWorkflows(noOpTracker{})

	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 2, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 3, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 2, len(buildRunResults.SkippedSteps))

	require.Equal(t, "1", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "1", os.Getenv("STEPLIB_BUILD_STATUS"))
}

// Trivial test for workflow environment handling, before workflows env should be visible in target and after workflow
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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))

	require.Equal(t, "0", os.Getenv("BITRISE_BUILD_STATUS"))
	require.Equal(t, "0", os.Getenv("STEPLIB_BUILD_STATUS"))
}

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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["target"]
	require.Equal(t, true, found)

	if os.Getenv("working_dir") != "" {
		require.Equal(t, nil, os.Unsetenv("working_dir"))
	}

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
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

	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	_, found := config.Workflows["out-test"]
	require.Equal(t, true, found)

	_, err = config.Validate()
	require.NoError(t, err)

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "out-test"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
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

func TestExpandEnvs(t *testing.T) {
	configStr := `
format_version: 1.3.0
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

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "test"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.NoError(t, err)
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestEvaluateInputs(t *testing.T) {
	configStr := `
format_version: 1.3.0
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

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "test"}
	runner := NewWorkflowRunner(runConfig)
	buildRunResults, err := runner.runWorkflows(noOpTracker{})
	require.Equal(t, nil, err)
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
	require.Equal(t, 2, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
}

func TestInvalidStepID(t *testing.T) {
	configStr := `
format_version: 1.3.0
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

	require.NoError(t, configs.InitPaths())

	runConfig := RunConfig{Config: config, Workflow: "target"}
	runner := NewWorkflowRunner(runConfig)
	results, err := runner.runWorkflows(noOpTracker{})
	require.Equal(t, 1, len(results.StepmanUpdates))
}

func TestPluginTriggered(t *testing.T) {
	bitriseYML := `
  format_version: 1.3.0
  default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"
  
  workflows:
    test:
      steps:
      - script:
          inputs:
          - content: |
              #!/bin/bash
              echo "test"
  `

	pluginYMFormat := `
name: testplugin
description: |-
  Manage Bitrise CLI steps
%s
executable:
  osx: {executable_url}
  linux: {executable_url}
`

	pluginSpecYMLFormat := `
route_map:
  testplugin:
    name: testplugin
    source: https://whatever.com
    version: 1.3.0
    commit_hash: ""
    %s
    latest_available_version: ""
`

	type testCase struct {
		name                    string
		pluginTrigger           string
		specTrigger             string
		expectedTriggeredEvents []string
	}

	testCases := []testCase{
		{
			"GivenPluginRegisteredForTrigger_ThenPluginTriggeredOnce",
			"trigger: DidFinishRun",
			"trigger: DidFinishRun",
			[]string{`"event_name":"DidFinishRun"`},
		},
		{
			"GivenPluginRegisteredForSingleTriggers_ThenPluginTriggeredOnce",
			"triggers:\n  - DidFinishRun",
			"triggers:\n      - DidFinishRun",
			[]string{`"event_name":"DidFinishRun"`},
		},
		{
			"GivenPluginRegisteredForMultipleTriggers_ThenPluginTriggeredTwice",
			"triggers:\n  - WillStartRun\n  - DidFinishRun",
			"triggers:\n      - WillStartRun\n      - DidFinishRun",
			[]string{`"event_name":"WillStartRun"`, `"event_name":"DidFinishRun"`},
		},
		{
			"GivenPluginRegisteredForMultipleTriggers_ThenPluginTriggeredTwice",
			"trigger: WillStartRun\ntriggers:\n  - DidFinishRun",
			"trigger: WillStartRun\n    triggers:\n      - DidFinishRun",
			[]string{`"event_name":"WillStartRun"`, `"event_name":"DidFinishRun"`},
		},
	}

	for _, test := range testCases {
		t.Log(test.name)
		{
			// Given
			config := givenWorkflowLoaded(t, bitriseYML)
			pluginYML := fmt.Sprintf(pluginYMFormat, test.pluginTrigger)
			pluginSpec := fmt.Sprintf(pluginSpecYMLFormat, test.specTrigger)
			givenPluginInstalled(t, pluginYML, "testplugin", pluginSpec)

			// When
			var origWiter io.Writer
			var buf bytes.Buffer
			opts := log.GetGlobalLoggerOpts()
			origWiter = opts.Writer
			opts.Writer = &buf
			log.InitGlobalLogger(opts)

			runConfig := RunConfig{Config: config, Workflow: "test"}
			runner := NewWorkflowRunner(runConfig)
			_, err := runner.runWorkflows(noOpTracker{})
			opts.Writer = origWiter

			// Then
			require.NoError(t, err)
			for _, expectedEvent := range test.expectedTriggeredEvents {
				condition := func() bool {
					output := buf.String()
					return strings.Contains(output, expectedEvent)
				}
				assert.Eventuallyf(t, condition, 5*time.Second, 150*time.Millisecond, "", "")
			}
		}
	}
}

func TestIsSecretFiltering(t *testing.T) {
	t.Log("flag influences the filtering")
	{
		set, err := isSecretFiltering(pointers.NewBoolPtr(true), []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.True(t, set)

		set, err = isSecretFiltering(pointers.NewBoolPtr(false), []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.False(t, set)
	}

	t.Log("secret influences the filtering")
	{
		set, err := isSecretFiltering(nil, []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "true"},
		})
		require.NoError(t, err)
		require.True(t, set)

		set, err = isSecretFiltering(nil, []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "false"},
		})
		require.NoError(t, err)
		require.False(t, set)
	}

	t.Log("flag has priority")
	{
		set, err := isSecretFiltering(pointers.NewBoolPtr(false), []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "true"},
		})
		require.NoError(t, err)
		require.False(t, set)
	}

	t.Log("secrets are exposed")
	{
		set, err := isSecretFiltering(nil, []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "true", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "false", "opts": map[string]interface{}{"is_expand": true}},
		})
		require.NoError(t, err)
		require.False(t, set)

		set, err = isSecretFiltering(nil, []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "true", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"BITRISE_SECRET_FILTERING": "$BITRISE_SECRET_FILTERING", "opts": map[string]interface{}{"is_expand": true}},
		})
		require.NoError(t, err)
		require.True(t, set)
	}
}

func TestIsPRMode(t *testing.T) {
	t.Run("Should be false for: prGlobalFlag: nil, prModeEnv: '', prIDEnv: ''", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, false, pr)
	})

	t.Run("Should be true for: prGlobalFlag: true, prModeEnv: '', prIDEnv: ''", func(t *testing.T) {
		t.Setenv(configs.PRModeEnvKey, "")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(pointers.NewBoolPtr(true), []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be false for: prGlobalFlag: nil, prModeEnv: '', prIDEnv: '', secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "false"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, false, pr)
	})

	t.Run("Should be false for: prGlobalFlag: nil, prModeEnv: 'false', prIDEnv: '', secrets: ''", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: ""
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, false, pr)
	})

	t.Run("Should be false for: prGlobalFlag: false, prModeEnv: 'true', prIDEnv: 'ID', secrets: 'true'", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "true"
- PULL_REQUEST_ID: "ID"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "true")
		t.Setenv(configs.PullRequestIDEnvKey, "ID")

		pr, err := isPRMode(pointers.NewBoolPtr(false), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, false, pr)
	})

	t.Run("Should be true for: prGlobalFlag: true, prModeEnv: '', prIDEnv: '', secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "false"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(pointers.NewBoolPtr(true), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be true for: prGlobalFlag: nil, prModeEnv: 'true', prIDEnv: '', secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "false"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "true")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be true for: prGlobalFlag: nil, prModeEnv: 'false', prIDEnv: 'some', secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "false"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "some")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be true for: prGlobalFlag: nil, prModeEnv: '', prIDEnv: '', secrets: true", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "true"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be true for: prGlobalFlag: nil, prModeEnv: 'false', prIDEnv: '', secrets: true", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: ""
- PULL_REQUEST_ID: "some"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})

	t.Run("Should be true for: prGlobalFlag: true, prModeEnv: 'false', prIDEnv: '', secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- PR: "false"
- PULL_REQUEST_ID: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.PRModeEnvKey, "false")
		t.Setenv(configs.PullRequestIDEnvKey, "")

		pr, err := isPRMode(pointers.NewBoolPtr(true), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})
}

func TestIsCIMode(t *testing.T) {
	t.Run("Should be false for: ciGlobalFlag: nil, ciModeEnv: 'false'", func(t *testing.T) {
		t.Setenv(configs.CIModeEnvKey, "false")

		ci, err := isCIMode(nil, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, false, ci)
	})

	t.Run("Should be true for: ciGlobalFlag: true, ciModeEnv: 'false'", func(t *testing.T) {
		t.Setenv(configs.CIModeEnvKey, "")

		ci, err := isCIMode(pointers.NewBoolPtr(true), []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, true, ci)
	})

	t.Run("Should be false for: ciGlobalFlag: false, ciModeEnv: 'false' secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- CI: "false"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.CIModeEnvKey, "false")

		ci, err := isCIMode(pointers.NewBoolPtr(false), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, false, ci)
	})

	t.Run("Should be true for: ciGlobalFlag: true, ciModeEnv: '' secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- CI: "false"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.CIModeEnvKey, "")

		ci, err := isCIMode(pointers.NewBoolPtr(true), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, ci)
	})

	t.Run("Should be true for: ciGlobalFlag: nil, ciModeEnv: 'true' secrets: false", func(t *testing.T) {
		inventoryStr := `
envs:
- CI: ""
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.CIModeEnvKey, "true")

		ci, err := isCIMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, ci)
	})

	t.Run("Should be true for: ciGlobalFlag: nil, ciModeEnv: '' secrets: true", func(t *testing.T) {
		inventoryStr := `
envs:
- CI: "true"
`
		inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
		require.NoError(t, err)

		t.Setenv(configs.CIModeEnvKey, "")

		ci, err := isCIMode(nil, inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, ci)
	})
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

func givenWorkflowLoaded(t *testing.T, workflow string) models.BitriseDataModel {
	require.NoError(t, configs.InitPaths())
	config, warnings, err := bitrise.ConfigModelFromYAMLBytes([]byte(workflow))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	return config
}

func givenPluginInstalled(t *testing.T, pluginContent, pluginName, pluginSpec string) {
	bitrisePath := givenPlugin(t, pluginContent, pluginName, pluginSpec)
	plugins.ForceInitPaths(bitrisePath)
}

func givenPlugin(t *testing.T, pluginContent, pluginName, pluginSpec string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__plugin_test__")
	require.NoError(t, err)

	bitriseDir := filepath.Join(tmpDir, ".bitrise")
	pluginsDir := filepath.Join(bitriseDir, "plugins")
	pluginSrcDir := filepath.Join(pluginsDir, pluginName, "src")

	// Create bitrise-plugin.sh
	pluginScriptPth := filepath.Join(pluginSrcDir, "bitrise-plugin.sh")
	pluginSHContent := fmt.Sprintf(`
  #!/bin/bash

  cat /dev/stdin

  echo "%s-called"
  `, pluginName)
	write(t, pluginSHContent, pluginScriptPth)
	err = os.Chmod(pluginScriptPth, 0777)
	require.NoError(t, err)

	// Create bitrise-plugin.yml
	pluginYMLContent := strings.ReplaceAll(pluginContent, "{executable_url}", "file://"+pluginScriptPth)
	pluginYMLPth := filepath.Join(pluginSrcDir, "bitrise-plugin.yml")
	write(t, pluginYMLContent, pluginYMLPth)

	// Create spec.yml
	specYMLContent := strings.ReplaceAll(pluginSpec, "{executable_url}", "file://"+pluginScriptPth)
	specYMLPth := filepath.Join(pluginsDir, "spec.yml")
	write(t, specYMLContent, specYMLPth)

	return bitriseDir
}

func write(t *testing.T, content, toPth string) {
	toDir := filepath.Dir(toPth)
	exist, err := pathutil.IsDirExists(toDir)
	require.NoError(t, err)
	if !exist {
		require.NoError(t, os.MkdirAll(toDir, 0700))
	}
	require.NoError(t, fileutil.WriteStringToFile(toPth, content))
}

type noOpTracker struct{}

func (n noOpTracker) SendStepStartedEvent(analytics.Properties, cliAnalytics.StepInfo, map[string]interface{}, map[string]string) {
}
func (n noOpTracker) SendStepFinishedEvent(analytics.Properties, cliAnalytics.StepResult) {}
func (n noOpTracker) SendCLIWarning(string)                                               {}
func (n noOpTracker) SendWorkflowStarted(analytics.Properties, string, string)            {}
func (n noOpTracker) SendWorkflowFinished(analytics.Properties, bool)                     {}
func (n noOpTracker) Wait()                                                               {}
