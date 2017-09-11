package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestEnvListSizeInBytes(t *testing.T) {
	str100Bytes := strings.Repeat("a", 100)
	require.Equal(t, 100, len([]byte(str100Bytes)))

	env := models.EnvironmentItemModel{
		"key": str100Bytes,
	}

	envList := []models.EnvironmentItemModel{env}
	size, err := envListSizeInBytes(envList)
	require.Equal(t, nil, err)
	require.Equal(t, 100, size)

	envList = []models.EnvironmentItemModel{env, env}
	size, err = envListSizeInBytes(envList)
	require.Equal(t, nil, err)
	require.Equal(t, 200, size)
}

func TestValidateEnv(t *testing.T) {
	// Valid - max allowed
	str20KBytes := strings.Repeat("a", (20 * 1024))
	env1 := models.EnvironmentItemModel{
		"key": str20KBytes,
	}
	envs := []models.EnvironmentItemModel{env1}

	valValue, err := validateEnv("key", str20KBytes, envs)
	require.NoError(t, err)
	require.Equal(t, str20KBytes, valValue)

	// List oversize
	//  first create a large, but valid env set
	for i := 0; i < 3; i++ {
		envs = append(envs, env1)
	}

	valValue, err = validateEnv("key", str20KBytes, envs)
	require.NoError(t, err)
	require.Equal(t, str20KBytes, valValue)

	// append one more -> too large
	envs = append(envs, env1)
	_, err = validateEnv("key", str20KBytes, envs)
	require.Equal(t, errors.New("environment list too large"), err)

	// List oversize + too big value
	str10Kbytes := strings.Repeat("a", (10 * 1024))
	env1 = models.EnvironmentItemModel{
		"key": str10Kbytes,
	}
	envs = []models.EnvironmentItemModel{}
	for i := 0; i < 8; i++ {
		env := models.EnvironmentItemModel{
			"key": str10Kbytes,
		}
		envs = append(envs, env)
	}

	str21Kbytes := strings.Repeat("a", (21 * 1024))

	valValue, err = validateEnv("key", str21Kbytes, envs)
	require.NoError(t, err)
	require.Equal(t, "environment value too large - rejected", valValue)
}
