package cli

import (
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/bitrise"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
)

func Test0Steps1Workflows(t *testing.T) {
	workflow := models.WorkflowModel{}

	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

	config := models.BitriseDataModel{
		FormatVersion:        "1.0.0",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		Workflows: map[string]models.WorkflowModel{
			"zero_steps": models.WorkflowModel{},
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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

func Test0Steps3WorkflowsBeforeAfter(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

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
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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

func Test0Steps3WorkflowsCircularDependency(t *testing.T) {
	if err := os.Setenv("BITRISE_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", "0"); err != nil {
		t.Fatal(err)
	}

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
		t.Fatal("No workflow found with title (trivial_fail)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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
          title: Should skipped
  `
	config, err := bitrise.ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}
	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with title (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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
		t.Fatal("Should found workflow reference cycle")
	}
}

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
	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with title (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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
	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with title (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
	t.Log("Buil run results:", buildRunResults)
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
	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with title (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}
	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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
	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("No workflow found with title (target)")
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}

	buildRunResults := models.BuildRunResultsModel{
		StartTime: time.Now(),
	}

	buildRunResults = activateAndRunWorkflow(workflow, config, buildRunResults)
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
