package cli

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestMerge(t *testing.T) {
	root := `format_version: "13"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
include:
  - path: included.yml
project_type: android
meta:
  bitrise.io:
    stack: linux-docker-android-20.04
    machine_type_id: standard
workflows:
  ui_test_on_tablet:
    steps:
    - pull-intermediate-files@1:
        inputs:
        - artifact_sources: build_tests.build_for_ui_testing
  ui_test_on_foldable:
    envs:
    - EMULATOR_PROFILE: 8in Foldable
    before_run:
    - _pull_apks
    after_run:
    - _run_tests`

	included := `workflows:
  ui_test_on_phone:
    envs:
    - EMULATOR_PROFILE: pixel_5
    before_run:
    - _pull_apks
    after_run:
    - _run_tests
  ui_test_on_tablet:
    envs:
    - EMULATOR_PROFILE: 10.1in WXGA (Tablet)
    before_run:
    - _pull_apks
    after_run:
    - _run_tests`

	expected := `
include:
  - path: included.yml
format_version: "13"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
project_type: android
meta:
  bitrise.io:
    stack: linux-docker-android-20.04
    machine_type_id: standard
workflows:
  ui_test_on_phone:
    envs:
    - EMULATOR_PROFILE: pixel_5
    before_run:
    - _pull_apks
    after_run:
    - _run_tests
  ui_test_on_tablet:
    steps:
    - pull-intermediate-files@1:
        inputs:
        - artifact_sources: build_tests.build_for_ui_testing
    envs:
    - EMULATOR_PROFILE: 10.1in WXGA (Tablet)
    before_run:
    - _pull_apks
    after_run:
    - _run_tests
  ui_test_on_foldable:
    envs:
    - EMULATOR_PROFILE: 8in Foldable
    before_run:
    - _pull_apks
    after_run:
    - _run_tests
`

	ymlTree := YmlTreeModel{
		Config: root,
		Includes: []YmlTreeModel{
			{Config: included},
		},
	}

	result, err := runMerge(&ymlTree)
	require.NoError(t, err)

	resultString, err := yaml.Marshal(result)

	require.YAMLEq(t, expected, string(resultString))
}
