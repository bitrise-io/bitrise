package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestStepOutputsInTemplate(t *testing.T) {
	inventoryStr := `
envs:
- TEMPLATE_TEST0: "true"
`
	inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)
	require.Equal(t, 5, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 1, len(buildRunResults.SkippedSteps))
}

func TestFailedStepOutputs(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "test", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, nil, err)
	require.Equal(t, 1, len(buildRunResults.SuccessSteps))
	require.Equal(t, 0, len(buildRunResults.FailedSteps))
	require.Equal(t, 1, len(buildRunResults.FailedSkippableSteps))
	require.Equal(t, 0, len(buildRunResults.SkippedSteps))
}

func TestBitriseSourceDir(t *testing.T) {
	currPth, err := pathutil.NormalizedOSTempDirPath("bitrise_source_dir_test")
	if err != nil {
		t.Fatal("Failed to get curr abs path: ", err)
	}

	testPths := []string{}
	for i := 0; i < 4; i++ {
		testPth := path.Join(currPth, fmt.Sprintf("_test%d", i))
		if err := os.RemoveAll(testPth); err != nil {
			t.Fatalf("Failed to remove %s, err: %s: ", testPth, err)
		}

		err := os.Mkdir(testPth, 0777)
		if err != nil {
			t.Fatalf("Failed to create %s, err: %s: ", testPth, err)
		}
		// eval symlinks: the Go generated temp folder on OS X is a symlink
		//  from /var/ to /private/var/
		testPth, err = filepath.EvalSymlinks(testPth)
		if err != nil {
			t.Fatalf("Failed to EvalSymlinks for (path:%s), err: %s", testPth, err)
		}

		defer func() {
			err := os.RemoveAll(testPth)
			if err != nil {
				t.Fatalf("Failed to remove %s, err: %s: ", testPth, err)
			}
		}()

		testPths = append(testPths, testPth)
	}

	//
	// BITRISE_SOURCE_DIR defined in Secret
	inventoryStr := `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
	inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// BITRISE_SOURCE_DIR defined in Secret, and in App
	inventoryStr = `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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
	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// BITRISE_SOURCE_DIR defined in Secret, App and Workflow
	inventoryStr = `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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
	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// BITRISE_SOURCE_DIR defined in Secret, App and Workflow
	//  BUT the value is empty in Workflow and App Envs - Secrets should be used!
	inventoryStr = `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - BITRISE_SOURCE_DIR:

workflows:
  test:
    envs:
    - BITRISE_SOURCE_DIR: ""
    steps:
    - script:
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo "BITRISE_SOURCE_DIR: $BITRISE_SOURCE_DIR"
            if [[ "$BITRISE_SOURCE_DIR" != "` + testPths[0] + `" ]] ; then
              echo "-> BITRISE_SOURCE_DIR missmatch!"
              exit 1
            fi
            curr_pwd="$(pwd)"
            if [[ "${curr_pwd}" != "` + testPths[0] + `" ]] ; then
              echo "-> pwd missmatch! : curr_pwd : ${curr_pwd}"
              exit 1
            fi
`
	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// BITRISE_SOURCE_DIR defined in secret, App, Workflow and Step
	inventoryStr = `
envs:
- BITRISE_SOURCE_DIR: "` + testPths[0] + `"
`
	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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
	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)
}

func TestEnvOrders(t *testing.T) {
	//
	// Only secret env - secret env should be use
	inventoryStr := `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

	inventory, err := bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr := `
format_version: 1.0.0
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

	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// Secret env & app env also defined - app env should be use
	inventoryStr = `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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

	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// Secret env & app env && workflow env also defined - workflow env should be use
	inventoryStr = `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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

	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)

	//
	// Secret env & app env && workflow env && step output env also defined - step output env should be use
	inventoryStr = `
envs:
- ENV_ORDER_TEST: "should be the 1."
`

	inventory, err = bitrise.InventoryModelFromYAMLBytes([]byte(inventoryStr))
	require.Equal(t, nil, err)

	configStr = `
format_version: 1.0.0
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

	config, err = bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)

	_, err = runWorkflowWithConfiguration(time.Now(), "test", config, inventory.Envs)
	require.Equal(t, nil, err)
}

// Test - Bitrise activateAndRunWorkflow
// If workflow contains no steps
func Test0Steps1Workflows(t *testing.T) {
	workflow := models.WorkflowModel{}

	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("BITRISE_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("STEPLIB_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"zero_steps": workflow,
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}
	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "zero_steps", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	if len(buildRunResults.SuccessSteps) != 0 {
		t.Fatalf("Success step count (%d), should be (0)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise activateAndRunWorkflow
// Workflow contains before and after workflow, and no one contains steps
func Test0Steps3WorkflowsBeforeAfter(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("BITRISE_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("STEPLIB_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

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

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}
	var err error
	buildRunResults, err = activateAndRunWorkflow("target", workflow, config, buildRunResults, &[]envmanModels.EnvironmentItemModel{}, "")
	if err != nil {
		t.Fatal("Failed to activate and run worfklow:", err)
	}
	if len(buildRunResults.SuccessSteps) != 0 {
		t.Fatalf("Success step count (%d), should be (0)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise Validate workflow
// Workflow contains before and after workflow, and no one contains steps, but circular wofklow dependecy exist, which should fail
func Test0Steps3WorkflowsCircularDependency(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("BITRISE_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// env cleanup
		if err := os.Unsetenv("STEPLIB_BUILD_STATUS"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

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

	if err := config.Validate(); err == nil {
		t.Fatal("Circular dependency, should fail")
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise activateAndRunWorkflow
// Trivial test with 1 workflow
func Test1Workflows(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	workflow, found := config.Workflows["trivial_fail"]
	if !found {
		t.Fatal("No workflow found with ID (trivial_fail)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime:      time.Now(),
		StepmanUpdates: map[string]int{},
	}
	buildRunResults, err = activateAndRunWorkflow("trivial_fail", workflow, config, buildRunResults, &[]envmanModels.EnvironmentItemModel{}, "")
	if err != nil {
		t.Fatal("Failed to activate and run worfklow:", err)
	}
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 1 {
		t.Fatalf("FailedSkippable step count (%d), should be (1)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise activateAndRunWorkflow
// Trivial test with before, after workflows
func Test3Workflows(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 2 {
		t.Fatalf("Failed step count (%d), should be (2)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 1 {
		t.Fatalf("FailedSkippable step count (%d), should be (1)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise ConfigModelFromYAMLBytes
// Workflow contains before and after workflow, and no one contains steps, but circular wofklow dependecy exist, which should fail
func TestRefeneceCycle(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	_, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err == nil {
		t.Fatal("Should find workflow reference cycle")
	}
}

// Test - Bitrise BuildStatusEnv
// Checks if BuildStatusEnv is set correctly
func TestBuildStatusEnv(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run results: %#v\n", buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 1 {
		t.Fatalf("FailedSkippable step count (%d), should be (1)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise activateAndRunWorkflow
// Trivial fail test
func TestFail(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Log("Build run results:", buildRunResults)
	if len(buildRunResults.SuccessSteps) != 3 {
		t.Fatalf("Success step count (%d), should be (3)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 1 {
		t.Fatalf("FailedSkippable step count (%d), should be (1)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 1 {
		t.Fatalf("Skipped step count (%d), should be (1)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise activateAndRunWorkflow
// Trivial success test
func TestSuccess(t *testing.T) {
	configStr := `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  target:
    steps:
    - script:
        title: Should success
    `
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise BuildStatusEnv
// Checks if BuildStatusEnv is set correctly
func TestBuildFailedMode(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run result: %#v", buildRunResults)
	if len(buildRunResults.SkippedSteps) != 2 {
		t.Fatalf("Skipped step count (%d), should be (2)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 1 {
		t.Fatalf("Failed step count (%d), should be (1)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "1" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "1" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise Environments
// Trivial test for workflow environment handling
// Before workflows env should be visible in target and after workflow
func TestWorkflowEnvironments(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run result: %#v", buildRunResults)
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 2 {
		t.Fatalf("Success step count (%d), should be (2)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise Environments
// Test for same env in before and target workflow, actual workflow should overwrite environemnt and use own value
func TestWorkflowEnvironmentOverWrite(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run result: %#v", buildRunResults)
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 2 {
		t.Fatalf("Success step count (%d), should be (2)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise Environments
// Target workflows env should be visible in before and after workflow
func TestTargetDefinedWorkflowEnvironment(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run result: %#v", buildRunResults)
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 1 {
		t.Fatalf("Success step count (%d), should be (1)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Test - Bitrise Environments
// Step input should visible only for actual step and invisible for other steps
func TestStepInputEnvironment(t *testing.T) {
	configStr := `
format_version: 1.0.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  before:
    steps:
    - script:
        inputs:
        - working_dir: $HOME

  target:
    title: target
    before_run:
    - before
    steps:
    - script:
        title: "${working_dir} should not exist"
        inputs:
        - content: |
            #!/bin/bash
            set -v
            echo ${ENV}
            if [ ! -z "$working_dir" ] ; then
              echo ${working_dir}
              exit 3
            fi
`
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	_, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with ID (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("working_dir") != "" {
		require.Equal(t, nil, os.Unsetenv("working_dir"))
	}

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "target", config, []envmanModels.EnvironmentItemModel{})
	t.Log("Err: ", err)
	t.Logf("Build run result: %#v", buildRunResults)
	if len(buildRunResults.SkippedSteps) != 0 {
		t.Fatalf("Skipped step count (%d), should be (0)", len(buildRunResults.SkippedSteps))
	}
	if len(buildRunResults.SuccessSteps) != 2 {
		t.Fatalf("Success step count (%d), should be (2)", len(buildRunResults.SuccessSteps))
	}
	if len(buildRunResults.FailedSteps) != 0 {
		t.Fatalf("Failed step count (%d), should be (0)", len(buildRunResults.FailedSteps))
	}
	if len(buildRunResults.FailedSkippableSteps) != 0 {
		t.Fatalf("FailedSkippable step count (%d), should be (0)", len(buildRunResults.FailedSkippableSteps))
	}

	if status := os.Getenv("BITRISE_BUILD_STATUS"); status != "0" {
		t.Log("BITRISE_BUILD_STATUS:", status)
		t.Fatal("BUILD_STATUS envs are incorrect")
	}
	if status := os.Getenv("STEPLIB_BUILD_STATUS"); status != "0" {
		t.Log("STEPLIB_BUILD_STATUS:", status)
		t.Fatal("STEPLIB_BUILD_STATUS envs are incorrect")
	}
}

// Outputs exported with `envman add` should be accessible for subsequent Steps.
func TestStepOutputEnvironment(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)
	_, found := config.Workflows["out-test"]
	require.Equal(t, true, found)

	err = config.Validate()
	require.Equal(t, nil, err)

	buildRunResults, err := runWorkflowWithConfiguration(time.Now(), "out-test", config, []envmanModels.EnvironmentItemModel{})
	require.Equal(t, "[BITRISE_CLI] - Workflow FINISHED but a couple of steps failed - Ouch", err.Error())
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
format_version: 1.0.0
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
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}

	last, err := lastWorkflowIDInConfig("target", config)
	if err != nil {
		t.Fatal(err)
	}

	if last != "after2" {
		t.Fatalf("Last workflow id is incorrect: (%s) should be (after2)", last)
	}
}
