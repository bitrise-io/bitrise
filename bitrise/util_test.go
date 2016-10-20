package bitrise

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/configs"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

func secToDuration(sec float64) time.Duration {
	return time.Duration(sec * 1e9)
}

func minToDuration(min float64) time.Duration {
	return secToDuration(min * 60)
}

func hourToDuration(hour float64) time.Duration {
	return minToDuration(hour * 60)
}

func TestTimeToFormattedSeconds(t *testing.T) {
	t.Log("formatted print rounds")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(0.999))
		require.NoError(t, err)
		require.Equal(t, "1.00 sec", timeStr)
	}

	t.Log("sec < 1.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(0.111))
		require.NoError(t, err)
		require.Equal(t, "0.11 sec", timeStr)
	}

	t.Log("sec < 10.0")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(9.111))
		require.NoError(t, err)
		require.Equal(t, "9.11 sec", timeStr)
	}

	t.Log("sec < 600 | min < 10")
	{
		timeStr, err := FormattedSecondsToMax8Chars(secToDuration(599.111))
		require.NoError(t, err)
		require.Equal(t, "599 sec", timeStr)
	}

	t.Log("min < 60")
	{
		timeStr, err := FormattedSecondsToMax8Chars(minToDuration(59.111))
		require.NoError(t, err)
		require.Equal(t, "59.1 min", timeStr)
	}

	t.Log("hour < 10")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(9.111))
		require.NoError(t, err)
		require.Equal(t, "9.1 hour", timeStr)
	}

	t.Log("hour < 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(999.111))
		require.NoError(t, err)
		require.Equal(t, "999 hour", timeStr)
	}

	t.Log("hour >= 1000")
	{
		timeStr, err := FormattedSecondsToMax8Chars(hourToDuration(1000))
		require.EqualError(t, err, "time (1000.000000 hour) greater then max allowed (999 hour)")
		require.Equal(t, "", timeStr)
	}
}

func TestRemoveConfigRedundantFieldsAndFillStepOutputs(t *testing.T) {
	// setup
	require.NoError(t, configs.InitPaths())

	configStr := `
  format_version: 1.3.0
  default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

  workflows:
    remove_test:
      steps:
      - script:
          inputs:
          - content: |
              #!/bin/bash
              set -v
              exit 2
            opts:
              is_expand: true
      - timestamp:
          title: Generate timestamps
    `

	config, warnings, err := ConfigModelFromYAMLBytes([]byte(configStr))
	require.Equal(t, nil, err)
	require.Equal(t, 0, len(warnings))

	require.Equal(t, nil, RemoveConfigRedundantFieldsAndFillStepOutputs(&config))

	for workflowID, workflow := range config.Workflows {
		if workflowID == "remove_test" {
			for _, stepListItem := range workflow.Steps {
				for stepID, step := range stepListItem {
					if stepID == "script" {
						for _, input := range step.Inputs {
							key, _, err := input.GetKeyValuePair()
							require.Equal(t, nil, err)

							if key == "content" {
								opts, err := input.GetOptions()
								require.Equal(t, nil, err)

								// script content should keep is_expand: true, becouse it's diffenet from spec default
								require.Equal(t, true, *opts.IsExpand)
							}
						}
					} else if stepID == "timestamp" {
						// timestamp title should be nil, becouse it's the same as spec value
						require.Equal(t, (*string)(nil), step.Title)

						for _, output := range step.Outputs {
							key, _, err := output.GetKeyValuePair()
							require.Equal(t, nil, err)

							if key == "UNIX_TIMESTAMP" {
								// timestamp outputs should filled with key-value & opts.Title
								opts, err := output.GetOptions()
								require.Equal(t, nil, err)

								require.Equal(t, "unix style", *opts.Title)
								require.Equal(t, (*bool)(nil), opts.IsExpand)
								require.Equal(t, (*bool)(nil), opts.IsDontChangeValue)
								require.Equal(t, (*bool)(nil), opts.IsRequired)
							}
						}
					}
				}
			}
		}
	}

	// timestamp outputs should filled with key-value & opts.Title

}

func TestSsStringSliceWithSameElements(t *testing.T) {
	s1 := []string{}
	s2 := []string{}
	require.Equal(t, true, isStringSliceWithSameElements(s1, s2))

	s1 = []string{"1", "2", "3"}
	s2 = []string{"2", "1"}
	require.Equal(t, false, isStringSliceWithSameElements(s1, s2))

	s2 = append(s2, "3")
	require.Equal(t, true, isStringSliceWithSameElements(s1, s2))

	s2 = []string{"1,", "1,", "1"}
	require.Equal(t, false, isStringSliceWithSameElements(s1, s2))
}

func TestIsDependecyEqual(t *testing.T) {
	d1 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep"}
	d2 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep"}

	require.Equal(t, true, isDependecyEqual(d1, d2))

	d1 = stepmanModels.DependencyModel{Manager: "manager", Name: "dep1"}
	d2 = stepmanModels.DependencyModel{Manager: "manager", Name: "dep"}

	require.Equal(t, false, isDependecyEqual(d1, d2))

	d1 = stepmanModels.DependencyModel{Manager: "manager", Name: "dep"}
	d2 = stepmanModels.DependencyModel{Manager: "manager1", Name: "dep"}

	require.Equal(t, false, isDependecyEqual(d1, d2))
}

func TestContainsDependecy(t *testing.T) {
	d1 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep1"}
	d2 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep2"}
	d3 := stepmanModels.DependencyModel{Manager: "manager1", Name: "dep3"}

	m := map[stepmanModels.DependencyModel]bool{
		d1: false,
		d2: true,
	}

	require.Equal(t, true, containsDependecy(m, d1))

	require.Equal(t, false, containsDependecy(m, d3))
}

func TestIsDependencySliceWithSameElements(t *testing.T) {
	s1 := []stepmanModels.DependencyModel{}
	s2 := []stepmanModels.DependencyModel{}
	require.Equal(t, true, isDependencySliceWithSameElements(s1, s2))

	d1 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep1"}
	d2 := stepmanModels.DependencyModel{Manager: "manager", Name: "dep2"}
	d3 := stepmanModels.DependencyModel{Manager: "manager1", Name: "dep3"}

	s1 = []stepmanModels.DependencyModel{d1, d2, d3}
	s2 = []stepmanModels.DependencyModel{d2, d1}
	require.Equal(t, false, isDependencySliceWithSameElements(s1, s2))

	s2 = append(s2, d3)
	require.Equal(t, true, isDependencySliceWithSameElements(s1, s2))

	s2 = []stepmanModels.DependencyModel{d1, d1, d1}
	require.Equal(t, false, isDependencySliceWithSameElements(s1, s2))
}

func TestConfigModelFromYAMLBytes(t *testing.T) {
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
	config, warnings, err := ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	workflow, found := config.Workflows["trivial_fail"]
	require.Equal(t, true, found)
	require.Equal(t, 6, len(workflow.Steps))
}

func TestConfigModelFromJSONBytes(t *testing.T) {
	configStr := `
{
  "format_version": "1.0.0",
  "default_step_lib_source": "https://github.com/bitrise-io/bitrise-steplib.git",
  "app": {
    "envs": null
  },
  "workflows": {
    "trivial_fail": {
      "title": "",
      "summary": "",
      "before_run": null,
      "after_run": null,
      "envs": null,
      "steps": [
        {
          "script": {
            "title": "Should success",
            "source": {}
          }
        },
        {
          "script": {
            "title": "Should fail, but skippable",
            "source": {},
            "is_skippable": true,
            "inputs": [
              {
                "content": "#!/bin/bash\nset -v\nexit 2\n",
                "opts": {}
              }
            ]
          }
        },
        {
          "script": {
            "title": "Should success",
            "source": {}
          }
        },
        {
          "script": {
            "title": "Should fail",
            "source": {},
            "inputs": [
              {
                "content": "#!/bin/bash\nset -v\nexit 2\n",
                "opts": {}
              }
            ]
          }
        },
        {
          "script": {
            "title": "Should success",
            "source": {},
            "is_always_run": true
          }
        },
        {
          "script": {
            "title": "Should skipped",
            "source": {}
          }
        }
      ]
    }
  }
}
  `
	config, warnings, err := ConfigModelFromJSONBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	workflow, found := config.Workflows["trivial_fail"]
	require.Equal(t, true, found)
	require.Equal(t, 6, len(workflow.Steps))
}

func TestConfigModelFromYAMLBytesNormalize(t *testing.T) {
	configStr := `
format_version: 1.3.0
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

app:
  envs:
  - BITRISE_BIN_NAME: bitrise
    opts:
      is_expand: false
  - GITHUB_RELEASES_URL: https://github.com/bitrise-io/bitrise/releases
    opts:
      is_expand: false

workflows:
  test:
    steps:
    - script:
        title: Should fail, but skippable
        is_skippable: true
        inputs:
        - content: echo "Hello World"
          opts:
            is_expand: no
`
	config, warnings, err := ConfigModelFromYAMLBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	// should be able to serialize into JSON
	_, err = json.MarshalIndent(config, "", "\t")
	require.NoError(t, err)
}

func TestConfigModelFromJSONBytesNormalize(t *testing.T) {
	configStr := `
{
  "format_version": "1.0.0",
  "default_step_lib_source": "https://github.com/bitrise-io/bitrise-steplib.git",
  "app": {
    "envs": [
      {
        "BITRISE_BIN_NAME": "bitrise",
        "opts": {
          "is_expand": false
        }
      },
      {
        "GITHUB_RELEASES_URL": "https://github.com/bitrise-io/bitrise/releases",
        "opts": {
          "is_expand": false
        }
      }
    ]
  },
  "workflows": {
    "test": {
      "steps": [
        {
          "script": {
            "title": "Should fail, but skippable",
            "is_skippable": true,
            "inputs": [
              {
                "content": "echo \"Hello World\"",
                "opts": {
                  "is_expand": false
                }
              }
            ]
          }
        }
      ]
    }
  }
}
`
	config, warnings, err := ConfigModelFromJSONBytes([]byte(configStr))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))

	t.Log("The ConfigModelFromJSONBytes method should call the required Normalize methods, so that no map[interface{}]interface{} is left - which would prevent the JSON serialization.")
	t.Logf("Config: %#v", config)
	// should be able to serialize into JSON
	_, err = json.MarshalIndent(config, "", "\t")
	require.NoError(t, err)
}
