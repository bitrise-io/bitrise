package cli

import (
	"strings"
	"testing"

	"github.com/bitrise-io/envman/envman"
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
	defaultConfig, err := envman.GetConfigs()
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   string
		envList []models.EnvironmentItemModel
		wantErr error
	}{
		{
			name:  "Max allowed env var value",
			key:   "key",
			value: strings.Repeat("a", defaultConfig.EnvBytesLimitInKB*1024),
		},
		{
			name:    "Max allowed env var list",
			key:     "key",
			value:   strings.Repeat("a", defaultConfig.EnvListBytesLimitInKB/2*1024),
			envList: []models.EnvironmentItemModel{{"key": strings.Repeat("a", defaultConfig.EnvListBytesLimitInKB/2*1024)}},
		},
		{
			name:    "Too big env var value",
			key:     "key",
			value:   strings.Repeat("a", defaultConfig.EnvBytesLimitInKB*1024+1),
			wantErr: NewEnvVarValueTooLargeError("key", float64(defaultConfig.EnvBytesLimitInKB)+(1.0/1024.0), float64(defaultConfig.EnvBytesLimitInKB)),
		},
		{
			name:    "Too big env var list",
			key:     "key",
			value:   "a",
			envList: []models.EnvironmentItemModel{{"key": strings.Repeat("a", defaultConfig.EnvListBytesLimitInKB*1024)}},
			wantErr: NewEnvVarListTooLargeError(((float64)(defaultConfig.EnvListBytesLimitInKB))+(float64)(len("a"))/1024.0, float64(defaultConfig.EnvListBytesLimitInKB)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validValue, err := validateEnv(tt.key, tt.value, tt.envList)
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr, err)
				require.Equal(t, "", validValue)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.value, validValue)
			}
		})
	}
}
