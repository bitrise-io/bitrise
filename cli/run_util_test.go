package cli

import (
	"encoding/base64"
	"testing"
)

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
	t.Log("Config:", configBase64Str)

	config, err := GetBitriseConfigFromBase64Data(configBase64Str)
	if err != nil {
		t.Fatal("Failed to get config from base 64 data, err:", err)
	}

	if config.FormatVersion != "0.9.10" {
		t.Fatal("Invalid FormatVersion")
	}
	if config.DefaultStepLibSource != "https://github.com/bitrise-io/bitrise-steplib.git" {
		t.Fatal("Invalid FormatVersion")
	}

	workflow, found := config.Workflows["target"]
	if !found {
		t.Fatal("Failed to found workflow")
	}
	if workflow.Title != "target" {
		t.Fatal("Invalid workflow.Title")
	}
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
	t.Log("Inventory:", inventoryBase64Str)

	inventory, err := GetInventoryFromBase64Data(inventoryBase64Str)
	if err != nil {
		t.Fatal("Failed to get inventory from base 64 data, err:", err)
	}

	env := inventory[0]

	key, value, err := env.GetKeyValuePair()
	if err != nil {
		t.Fatal("Failed to get env key-value pair, err:", err)
	}

	if key != "MY_HOME" {
		t.Fatal("Invalid key")
	}
	if value != "$HOME" {
		t.Fatal("Invalid value")
	}

	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal("Failed to get env options, err:", err)
	}

	if *opts.IsExpand != true {
		t.Fatal("Invalid IsExpand")
	}
}
