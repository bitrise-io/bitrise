package cli

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestMerge_Success(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		ymlTree  YmlTreeModel
		expected string
	}{
		{
			name: "multiple included properties",
			ymlTree: YmlTreeModel{
				Config: "",
				Includes: []YmlTreeModel{
					{
						Config: `
property1: 1
property2: 2
property3: 3`},
					{
						Config: `
property3: 30
property4: 40
property5: 50`},
					{
						Config: `
property3: 300
property5: 500
property7: 700`},
				},
			},
			expected: `
property1: 1
property2: 2
property3: 300
property4: 40
property5: 500
property7: 700
`},
		{
			name: "multiple included lists",
			ymlTree: YmlTreeModel{
				Includes: []YmlTreeModel{
					{
						Config: `
list:
- item1
- item2`},
					{
						Config: `
list:
- item3
- item4`},
					{
						Config: `
list:
- item5
- item6`,
					},
				},
			},
			expected: `
list:
- item1
- item2
- item3
- item4
- item5
- item6
`},
		{
			name: "multiple included maps",
			ymlTree: YmlTreeModel{
				Includes: []YmlTreeModel{
					{
						Config: `
config:
    property1: 1
    property2: 2
    property3: 3`},
					{
						Config: `
config:
    property3: 30
    property4: 40
    property5: 50`},
					{
						Config: `
config:
    property3: 300
    property5: 500
    property7: 700`,
					},
				},
			},
			expected: `
config:
    property1: 1
    property2: 2
    property3: 300
    property4: 40
    property5: 500
    property7: 700
`},
		{
			name: "nested included properties",
			ymlTree: YmlTreeModel{
				Config: `
property1: 1
property2: 2
property3: 3`,
				Includes: []YmlTreeModel{
					{
						Config: `
property3: 30
property4: 40
property5: 50`,
						Includes: []YmlTreeModel{
							{
								Config: `
property3: 300
property5: 500
property7: 700`,
							},
						},
					},
				},
			},
			expected: `
property1: 1
property2: 2
property3: 3
property4: 40
property5: 50
property7: 700
`},
		{
			name: "nested included lists",
			ymlTree: YmlTreeModel{
				Config: `
list:
- item1
- item2`,
				Includes: []YmlTreeModel{
					{
						Config: `
list:
- item3
- item4`,
						Includes: []YmlTreeModel{
							{
								Config: `
list:
- item5
- item6`,
							},
						},
					},
				},
			},
			expected: `
list:
- item5
- item6
- item3
- item4
- item1
- item2
`},
		{
			name: "nested included maps",
			ymlTree: YmlTreeModel{
				Config: `
config:
    property1: 1
    property2: 2
    property3: 3`,
				Includes: []YmlTreeModel{
					{
						Config: `
config:
    property3: 30
    property4: 40
    property5: 50`,
						Includes: []YmlTreeModel{
							{
								Config: `
config:
    property3: 300
    property5: 500
    property7: 700`,
							},
						},
					},
				},
			},
			expected: `
config:
    property1: 1
    property2: 2
    property3: 3
    property4: 40
    property5: 50
    property7: 700
`},
		{
			name: "complex",
			ymlTree: YmlTreeModel{
				Config: `
simple: 1
map:
  map_simple: 2
  map_map:  
    map_map_simple: 3
    map_map_list:
    - a
    - b
    - c
  map_list:
    - d
    - e
    - f
list:
  - 4
  - list_map:
      list_map_1: 5
      list_map_2: 6
  - list_list:
    - 7
    - 8
`,
				Includes: []YmlTreeModel{
					{
						Config: `
simple: 100
another_simple: 200
map:
  extra: 42
list:
- more
- items
`,
						Includes: []YmlTreeModel{
							{
								Config: `
map:
  map_map:
    map_map_simple: 400
    map_map_list:
    - deep1
    - deep2
    map_map_deep: deep
`,
							},
						},
					},
					{
						Config: `
another_simple: 300
map:
  map_map:
    map_map_simple: 500
    map_map_another: 600
  map_list:
  - another_item
`,
					},
				},
			},
			expected: `
simple: 1
another_simple: 300
map:
  extra: 42
  map_simple: 2
  map_map:  
    map_map_simple: 3
    map_map_another: 600
    map_map_list:
    - deep1
    - deep2
    - a
    - b
    - c
    map_map_deep: deep
  map_list:
    - another_item
    - d
    - e
    - f
list:
  - more
  - items
  - 4
  - list_map:
      list_map_1: 5
      list_map_2: 6
  - list_list:
    - 7
    - 8
`},
		{
			name: "RFC example",
			ymlTree: YmlTreeModel{
				Config: `format_version: "13"
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
    - _run_tests`,
				Includes: []YmlTreeModel{
					{Config: `workflows:
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
    - _run_tests`},
				},
			},
			expected: `
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
`},
		{
			name: "mismatching types - simple into map",
			ymlTree: YmlTreeModel{
				Config: `item: value`,
				Includes: []YmlTreeModel{
					{
						Config: `
item:
  value1: 1
  value2: 2
`,
					},
				},
			},
			expected: `item: value
`,
		},
		{
			name: "mismatching types - map into simple",
			ymlTree: YmlTreeModel{
				Config: `
item:
  value1: 1
  value2: 2
`,
				Includes: []YmlTreeModel{
					{
						Config: `item: value`,
					},
				},
			},
			expected: `
item:
  value1: 1
  value2: 2
`,
		},
		{
			name: "mismatching types - simple into list",
			ymlTree: YmlTreeModel{
				Config: `item: value`,
				Includes: []YmlTreeModel{
					{
						Config: `
item:
  - value1
  - value2
`,
					},
				},
			},
			expected: `item: value
`,
		},
		{
			name: "mismatching types - list into simple",
			ymlTree: YmlTreeModel{
				Config: `
item:
  - value1
  - value2
`,
				Includes: []YmlTreeModel{
					{
						Config: `item: value`,
					},
				},
			},
			expected: `
item:
  - value1
  - value2
`,
		},
		{
			name: "mismatching types - map into list",
			ymlTree: YmlTreeModel{
				Config: `
item:
  value1: 1
  value2: 2`,
				Includes: []YmlTreeModel{
					{
						Config: `
item:
  - value1
  - value2
`,
					},
				},
			},
			expected: `item:
  value1: 1
  value2: 2
`,
		},
		{
			name: "mismatching types - list into map",
			ymlTree: YmlTreeModel{
				Config: `
item:
  - value1
  - value2
`,
				Includes: []YmlTreeModel{
					{
						Config: `item:
  value1: 1
  value2: 2`,
					},
				},
			},
			expected: `
item:
  - value1
  - value2
`,
		},
	} {
		test := test
		t.Run(
			test.name, func(t *testing.T) {
				t.Parallel()
				result, err := runMerge(&test.ymlTree)
				require.NoError(t, err)

				resultString, err := yaml.Marshal(result)

				require.YAMLEq(t, test.expected, string(resultString))
			})
	}
}

func TestMerge_Error(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name          string
		ymlTree       YmlTreeModel
		expectedError string
	}{
		{
			name: "parse error - invalid YML",
			ymlTree: YmlTreeModel{
				FileName: "bitrise.yml",
				Config:   `format_version: "13`,
			},
			expectedError: "failed to merge YML files, error: failed to parse YML file bitrise.yml, error: yaml: found unexpected end of stream",
		},
		{
			name: "parse error - tabs in YML",
			ymlTree: YmlTreeModel{
				FileName: "bitrise.yml",
				Includes: []YmlTreeModel{
					{
						FileName: "included.yml",
						Config: `
workflows:
	tabbed-item: invalid
`},
				},
			},
			expectedError: "failed to merge YML files, error: failed to merge YML file included.yml, error: failed to parse YML file included.yml, error: yaml: line 3: found character that cannot start any token",
		},
	} {
		test := test
		t.Run(
			test.name, func(t *testing.T) {
				t.Parallel()

				result, err := runMerge(&test.ymlTree)

				require.Nil(t, result)
				require.ErrorContains(t, err, test.expectedError)
			})
	}
}
