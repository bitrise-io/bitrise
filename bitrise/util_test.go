package bitrise

import (
	"encoding/json"
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestApplyOutputAliases(t *testing.T) {
	t.Log("apply alias on signle env")
	{
		envs := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"ORIGINAL_KEY": "value",
			},
		}

		outputEnvs := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"ORIGINAL_KEY": "ALIAS_KEY",
			},
		}

		updatedEnvs, err := ApplyOutputAliases(envs, outputEnvs)
		require.NoError(t, err)
		require.Equal(t, 1, len(updatedEnvs))

		updatedKey, value, err := updatedEnvs[0].GetKeyValuePair()
		require.Equal(t, "ALIAS_KEY", updatedKey)
		require.Equal(t, "value", value)
	}

	t.Log("apply alias on env list")
	{
		envs := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"ORIGINAL_KEY": "value",
			},
			envmanModels.EnvironmentItemModel{
				"SHOULD_NOT_CHANGE_KEY": "value",
			},
		}

		outputEnvs := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{
				"ORIGINAL_KEY": "ALIAS_KEY",
			},
		}

		updatedEnvs, err := ApplyOutputAliases(envs, outputEnvs)
		require.NoError(t, err)
		require.Equal(t, 2, len(updatedEnvs))

		{
			// this env should be updated
			updatedKey, value, err := updatedEnvs[0].GetKeyValuePair()
			require.NoError(t, err)
			require.Equal(t, "ALIAS_KEY", updatedKey)
			require.Equal(t, "value", value)
		}

		{
			// this env should NOT be updated
			key, value, err := updatedEnvs[1].GetKeyValuePair()
			require.NoError(t, err)
			require.Equal(t, "SHOULD_NOT_CHANGE_KEY", key)
			require.Equal(t, "value", value)
		}
	}
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

// Workflow contains before and after workflow, and no one contains steps, but circular workflow dependency exist, which should fail
func TestConfigModelFromYAMLBytesReferenceCycle(t *testing.T) {
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
	_, warnings, err := ConfigModelFromYAMLBytes([]byte(configStr))
	require.Error(t, err)
	require.Equal(t, 0, len(warnings))
}

func TestConfigModelFromYAMLFileContent_StepListValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name: "Invalid bitrise.yml: with group in a step bundle's steps list",
			config: `
format_version: '11'
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
step_bundles:
  test:
    steps:
    - with: {}`,
			wantErr: "failed to normalize step_bundle: 'with group' is not allowed in a step bundle's step list",
		},
		{
			name: "Invalid bitrise.yml: step bundle in a 'with group''s steps list",
			config: `
format_version: '11'
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
services:
  postgres:
    image: postgres:13
workflows:
  primary:
    steps:
    - with:
        services:
        - postgres
        steps:
        - bundle::test: {}`,
			wantErr: "invalid 'with group' in workflow (primary): step bundle is not allowed in a 'with group''s step list",
		},
		{
			name: "Invalid bitrise.yml: with group in a 'with group''s steps list",
			config: `
format_version: '11'
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
services:
  postgres:
    image: postgres:13
workflows:
  primary:
    steps:
    - with:
        services:
        - postgres
        steps:
        - with: {}`,
			wantErr: "invalid 'with group' in workflow (primary): 'with group' is not allowed in a 'with group''s step list",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ConfigModelFromFileContent([]byte(tt.config), false, ValidationTypeFull)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigModelFromJSONFileContent_StepListValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name: "Invalid bitrise.yml: with group in a step bundle's steps list",
			config: `{
  "format_version": "11",
  "default_step_lib_source": "https://github.com/bitrise-io/bitrise-steplib.git",
  "step_bundles": {
    "test": {
      "steps": [
        {
          "with": {}
        }
      ]
    }
  }
}`,
			wantErr: "failed to normalize step_bundle: 'with group' is not allowed in a step bundle's step list",
		},
		{
			name: "Invalid bitrise.yml: step bundle in a 'with group''s steps list",
			config: `{
  "format_version": "11",
  "default_step_lib_source": "https://github.com/bitrise-io/bitrise-steplib.git",
  "services": {
    "postgres": {
      "image": "postgres:13"
    }
  },
  "workflows": {
    "primary": {
      "steps": [
        {
          "with": {
            "services": [
              "postgres"
            ],
            "steps": [
              {
                "bundle::test": {}
              }
            ]
          }
        }
      ]
    }
  }
}`,
			wantErr: "invalid 'with group' in workflow (primary): step bundle is not allowed in a 'with group''s step list",
		},
		{
			name: "Invalid bitrise.yml: with group in a 'with group''s steps list",
			config: `{
  "format_version": "11",
  "default_step_lib_source": "https://github.com/bitrise-io/bitrise-steplib.git",
  "services": {
    "postgres": {
      "image": "postgres:13"
    }
  },
  "workflows": {
    "primary": {
      "steps": [
        {
          "with": {
            "services": [
              "postgres"
            ],
            "steps": [
              {
                "with": {}
              }
            ]
          }
        }
      ]
    }
  }
}`,
			wantErr: "invalid 'with group' in workflow (primary): 'with group' is not allowed in a 'with group''s step list",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ConfigModelFromFileContent([]byte(tt.config), true, ValidationTypeFull)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
