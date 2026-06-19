package cli

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/toolkits"
	"github.com/stretchr/testify/assert"
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
		setEnv(t, configs.PRModeEnvKey, "", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

		pr, err := isPRMode(nil, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, false, pr)
	})

	t.Run("Should be true for: prGlobalFlag: true, prModeEnv: '', prIDEnv: ''", func(t *testing.T) {
		setEnv(t, configs.PRModeEnvKey, "", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "false", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "true", true)
		setEnv(t, configs.PullRequestIDEnvKey, "ID", true)

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

		setEnv(t, configs.PRModeEnvKey, "", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "true", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "false", true)
		setEnv(t, configs.PullRequestIDEnvKey, "some", true)

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

		setEnv(t, configs.PRModeEnvKey, "", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "false", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

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

		setEnv(t, configs.PRModeEnvKey, "false", true)
		setEnv(t, configs.PullRequestIDEnvKey, "", true)

		pr, err := isPRMode(pointers.NewBoolPtr(true), inventory.Envs)
		require.NoError(t, err)
		require.Equal(t, true, pr)
	})
}

func TestIsCIMode(t *testing.T) {
	t.Run("Should be false for: ciGlobalFlag: nil, ciModeEnv: 'false'", func(t *testing.T) {
		setEnv(t, configs.CIModeEnvKey, "false", true)

		ci, err := isCIMode(nil, []envmanModels.EnvironmentItemModel{})
		require.NoError(t, err)
		require.Equal(t, false, ci)
	})

	t.Run("Should be true for: ciGlobalFlag: true, ciModeEnv: 'false'", func(t *testing.T) {
		setEnv(t, configs.CIModeEnvKey, "", true)

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

		setEnv(t, configs.CIModeEnvKey, "false", true)

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

		setEnv(t, configs.CIModeEnvKey, "", true)

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

		setEnv(t, configs.CIModeEnvKey, "true", true)

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

		setEnv(t, configs.CIModeEnvKey, "", true)

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

	config, warnings, err := GetBitriseConfigFromBase64Data(configBase64Str, bitrise.ValidationTypeFull)
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

type toolkitPrepareCall struct {
	stepExecutionID string
	toolkitName     string
	stepID          string
	stepVersion     string
	result          toolkits.PrepareForStepRunResult
	err             error
}

type toolkitCapturingTracker struct {
	noOpTracker
	calls []toolkitPrepareCall
}

func (c *toolkitCapturingTracker) SendToolkitPrepareEvent(stepExecutionID, toolkitName, stepID, stepVersion string, result toolkits.PrepareForStepRunResult, err error) {
	c.calls = append(c.calls, toolkitPrepareCall{
		stepExecutionID: stepExecutionID,
		toolkitName:     toolkitName,
		stepID:          stepID,
		stepVersion:     stepVersion,
		result:          result,
		err:             err,
	})
}

func TestTrackToolkitPrepare(t *testing.T) {
	tests := []struct {
		name         string
		result       toolkits.PrepareForStepRunResult
		err          error
		expectsEvent bool
	}{
		{
			name:         "bash no-op: zero duration, no error → no event",
			result:       toolkits.PrepareForStepRunResult{PrepareDuration: 0, CacheHit: false},
			err:          nil,
			expectsEvent: false,
		},
		{
			name:         "go compile: nonzero duration → event fired",
			result:       toolkits.PrepareForStepRunResult{PrepareDuration: 500 * time.Millisecond, CacheHit: false},
			err:          nil,
			expectsEvent: true,
		},
		{
			name:         "go cache hit: nonzero duration → event fired",
			result:       toolkits.PrepareForStepRunResult{PrepareDuration: 2 * time.Millisecond, CacheHit: true},
			err:          nil,
			expectsEvent: true,
		},
		{
			name:         "prepare error with zero duration → event still fired",
			result:       toolkits.PrepareForStepRunResult{PrepareDuration: 0, CacheHit: false},
			err:          errors.New("compile failed"),
			expectsEvent: true,
		},
		{
			name:         "prepare error with nonzero duration → event still fired",
			result:       toolkits.PrepareForStepRunResult{PrepareDuration: 200 * time.Millisecond, CacheHit: false},
			err:          errors.New("compile failed"),
			expectsEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &toolkitCapturingTracker{}
			sIDData := stepid.CanonicalID{IDorURI: "my-step", Version: "1.0.0"}
			trackToolkitPrepare(tracker, "exec-uuid", "go", sIDData, tt.result, tt.err)

			if tt.expectsEvent {
				assert.Len(t, tracker.calls, 1)
				assert.Equal(t, "exec-uuid", tracker.calls[0].stepExecutionID)
				assert.Equal(t, "go", tracker.calls[0].toolkitName)
				assert.Equal(t, "my-step", tracker.calls[0].stepID)
				assert.Equal(t, "1.0.0", tracker.calls[0].stepVersion)
				assert.Equal(t, tt.result, tracker.calls[0].result)
				assert.Equal(t, tt.err, tracker.calls[0].err)
			} else {
				assert.Empty(t, tracker.calls)
			}
		})
	}
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
