package cli

import (
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
	// Valid
	value := strings.Repeat("a", (20 * 1024))
	env1 := models.EnvironmentItemModel{
		"key": value,
	}
	envs := []models.EnvironmentItemModel{env1}

	require.Equal(t, nil, validateEnv("key", value, envs))

	// List oversize
	for i := 0; i < 4; i++ {
		env := models.EnvironmentItemModel{
			"key": value,
		}
		envs = append(envs, env)
	}

	require.NotEqual(t, nil, validateEnv("key", value, envs))

	// List oversize + to big value
	value = strings.Repeat("a", (10 * 1024))
	env1 = models.EnvironmentItemModel{
		"key": value,
	}
	envs = []models.EnvironmentItemModel{}
	for i := 0; i < 8; i++ {
		env := models.EnvironmentItemModel{
			"key": value,
		}
		envs = append(envs, env)
	}

	value2 := strings.Repeat("a", (21 * 1024))

	require.NotEqual(t, nil, validateEnv("key", value2, envs))
}
