package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestExpandEnvsInString(t *testing.T) {
	require.Equal(t, nil, os.Setenv("MY_ENV_KEY", "key"))

	inp := "${MY_ENV_KEY} of my home"
	expanded := expandEnvsInString(inp)

	key := os.Getenv("MY_ENV_KEY")
	require.NotEqual(t, "", expanded)
	require.Equal(t, key+" of my home", expanded)
}

func TestCommandEnvs(t *testing.T) {
	env1 := models.EnvironmentItemModel{
		"test_key1": "test_value1",
	}
	require.Equal(t, nil, env1.FillMissingDefaults())

	env2 := models.EnvironmentItemModel{
		"test_key2": "test_value2",
	}
	require.Equal(t, nil, env2.FillMissingDefaults())

	envs := []models.EnvironmentItemModel{env1, env2}

	sessionEnvs, err := commandEnvs(envs)
	require.Equal(t, nil, err)

	env1Found := false
	env2Found := false
	for _, envString := range sessionEnvs {
		comp := strings.Split(envString, "=")
		key := comp[0]
		value := comp[1]

		envKey1, envValue1, err := env1.GetKeyValuePair()
		require.Equal(t, nil, err)

		envKey2, envValue2, err := env2.GetKeyValuePair()
		require.Equal(t, nil, err)

		if key == envKey1 && value == envValue1 {
			env1Found = true
		}
		if key == envKey2 && value == envValue2 {
			env2Found = true
		}
	}
	require.Equal(t, true, env1Found)
	require.Equal(t, true, env2Found)

	// Test skip_if_empty

	// skip_if_empty=false && value=empty => should add
	env1 = models.EnvironmentItemModel{
		"test_key3": "",
	}
	require.Equal(t, nil, env1.FillMissingDefaults())

	env2 = models.EnvironmentItemModel{
		"test_key4": "test_value4",
	}
	require.Equal(t, nil, env2.FillMissingDefaults())

	envs = []models.EnvironmentItemModel{env1, env2}

	sessionEnvs, err = commandEnvs(envs)
	require.Equal(t, nil, err)

	env1Found = false
	env2Found = false
	for _, envString := range sessionEnvs {
		comp := strings.Split(envString, "=")
		key := comp[0]
		value := comp[1]

		envKey1, envValue1, err := env1.GetKeyValuePair()
		require.Equal(t, nil, err)

		envKey2, envValue2, err := env2.GetKeyValuePair()
		require.Equal(t, nil, err)

		if key == envKey1 && value == envValue1 {
			env1Found = true
		}
		if key == envKey2 && value == envValue2 {
			env2Found = true
		}
	}
	require.Equal(t, true, env1Found)
	require.Equal(t, true, env2Found)

	// skip_if_empty=true && value=empty => should NOT add
	env1 = models.EnvironmentItemModel{
		"test_key5": "",
		"opts": models.EnvironmentItemOptionsModel{
			SkipIfEmpty: pointers.NewBoolPtr(true),
		},
	}
	require.Equal(t, nil, env1.FillMissingDefaults())

	env2 = models.EnvironmentItemModel{
		"test_key6": "test_value6",
	}
	require.Equal(t, nil, env2.FillMissingDefaults())

	envs = []models.EnvironmentItemModel{env1, env2}

	sessionEnvs, err = commandEnvs(envs)
	require.Equal(t, nil, err)

	env1Found = false
	env2Found = false
	for _, envString := range sessionEnvs {
		comp := strings.Split(envString, "=")
		key := comp[0]
		value := comp[1]

		envKey1, envValue1, err := env1.GetKeyValuePair()
		require.Equal(t, nil, err)

		envKey2, envValue2, err := env2.GetKeyValuePair()
		require.Equal(t, nil, err)

		if key == envKey1 && value == envValue1 {
			env1Found = true
		}
		if key == envKey2 && value == envValue2 {
			env2Found = true
		}
	}
	require.Equal(t, false, env1Found)
	require.Equal(t, true, env2Found)

	// skip_if_empty=true && value=NOT_empty => should add
	env1 = models.EnvironmentItemModel{
		"test_key7": "test_value7",
		"opts": models.EnvironmentItemOptionsModel{
			SkipIfEmpty: pointers.NewBoolPtr(true),
		},
	}
	require.Equal(t, nil, env1.FillMissingDefaults())

	env2 = models.EnvironmentItemModel{
		"test_key8": "test_value8",
	}
	require.Equal(t, nil, env2.FillMissingDefaults())

	envs = []models.EnvironmentItemModel{env1, env2}

	sessionEnvs, err = commandEnvs(envs)
	require.Equal(t, nil, err)

	env1Found = false
	env2Found = false
	for _, envString := range sessionEnvs {
		comp := strings.Split(envString, "=")
		key := comp[0]
		value := comp[1]

		envKey1, envValue1, err := env1.GetKeyValuePair()
		require.Equal(t, nil, err)

		envKey2, envValue2, err := env2.GetKeyValuePair()
		require.Equal(t, nil, err)

		if key == envKey1 && value == envValue1 {
			env1Found = true
		}
		if key == envKey2 && value == envValue2 {
			env2Found = true
		}
	}
	require.Equal(t, true, env1Found)
	require.Equal(t, true, env2Found)
}
