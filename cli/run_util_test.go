package cli

import (
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

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

func TestAddTestMetadata(t *testing.T) {
	t.Log("test empty dir")
	{
		testDirPath, err := pathutil.NormalizedOSTempDirPath("testing")
		if err != nil {
			t.Fatalf("failed to create testing dir, error: %s", err)
		}

		testResultStepInfo := models.TestResultStepInfo{}

		exists, err := pathutil.IsDirExists(testDirPath)
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if !exists {
			t.Fatal("test dir should exits")
		}

		if err := addTestMetadata(testDirPath, testResultStepInfo); err != nil {
			t.Fatalf("failed to normalize test dir, error: %s", err)
		}

		exists, err = pathutil.IsDirExists(testDirPath)
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if exists {
			t.Fatal("test dir should not exits")
		}
	}

	t.Log("test not empty dir")
	{
		testDirPath, err := pathutil.NormalizedOSTempDirPath("testing")
		if err != nil {
			t.Fatalf("failed to create testing dir, error: %s", err)
		}

		testResultStepInfo := models.TestResultStepInfo{}

		exists, err := pathutil.IsDirExists(testDirPath)
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if !exists {
			t.Fatal("test dir should exits")
		}

		if err := fileutil.WriteStringToFile(filepath.Join(testDirPath, "test-file"), "test-content"); err != nil {
			t.Fatalf("failed to write file, error: %s", err)
		}

		if err := addTestMetadata(testDirPath, testResultStepInfo); err != nil {
			t.Fatalf("failed to normalize test dir, error: %s", err)
		}

		exists, err = pathutil.IsDirExists(testDirPath)
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if !exists {
			t.Fatal("test dir should exits")
		}

		exists, err = pathutil.IsPathExists(filepath.Join(testDirPath, "test-file"))
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if !exists {
			t.Fatal("test file should exits")
		}

		exists, err = pathutil.IsPathExists(filepath.Join(testDirPath, "step-info.json"))
		if err != nil {
			t.Fatalf("failed to check if dir exists, error: %s", err)
		}

		if !exists {
			t.Fatal("step-info.json file should exits")
		}
	}
}
