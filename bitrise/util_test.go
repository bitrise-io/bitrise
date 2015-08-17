package bitrise

import (
	"encoding/json"
	"testing"
)

func TestConfigModelFromYAMLBytes(t *testing.T) {
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
	config, err := ConfigModelFromYAMLBytes([]byte(configStr))
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
	if len(workflow.Steps) != 6 {
		t.Fatal("Not the expected Steps count")
	}
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
	config, err := ConfigModelFromJSONBytes([]byte(configStr))
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
	if len(workflow.Steps) != 6 {
		t.Fatal("Not the expected Steps count")
	}
}

func TestConfigModelFromYAMLBytesNormalize(t *testing.T) {
	configStr := `
format_version: 1.0.0
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
	config, err := ConfigModelFromYAMLBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("The ConfigModelFromYAMLBytes method should call the required Normalize methods, so that no map[interface{}]interface{} is left - which would prevent the JSON serialization.")
	t.Logf("Config: %#v", config)
	// should be able to serialize into JSON
	_, err = json.MarshalIndent(config, "", "\t")
	if err != nil {
		t.Fatalf("Failed to generate JSON: %s", err)
	}
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
	config, err := ConfigModelFromJSONBytes([]byte(configStr))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("The ConfigModelFromJSONBytes method should call the required Normalize methods, so that no map[interface{}]interface{} is left - which would prevent the JSON serialization.")
	t.Logf("Config: %#v", config)
	// should be able to serialize into JSON
	_, err = json.MarshalIndent(config, "", "\t")
	if err != nil {
		t.Fatalf("Failed to generate JSON: %s", err)
	}
}
