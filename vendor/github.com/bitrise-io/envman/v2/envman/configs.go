package envman

import (
	"encoding/json"
	"os"
	"path"
	"strconv"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	envmanConfigFileName         = "configs.json"
	defaultEnvBytesLimitInKB     = 256
	defaultEnvListBytesLimitInKB = 256

	// EnvBytesLimitInKBEnvKey and EnvListBytesLimitInKBEnvKey override the
	// per-value and list-size limits at the process level, taking precedence
	// over configs.json. This lets the limit be raised before the first
	// `envman add`, which a script step cannot do because the check runs while
	// preparing that step's own environment.
	EnvBytesLimitInKBEnvKey     = "ENVMAN_ENV_BYTES_LIMIT_IN_KB"
	EnvListBytesLimitInKBEnvKey = "ENVMAN_ENV_LIST_BYTES_LIMIT_IN_KB"
)

// ConfigsModel ...
type ConfigsModel struct {
	EnvBytesLimitInKB     int `json:"env_bytes_limit_in_kb,omitempty"`
	EnvListBytesLimitInKB int `json:"env_list_bytes_limit_in_kb,omitempty"`
}

func getEnvmanConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".envman")
}

func getEnvmanConfigsFilePath() string {
	return path.Join(getEnvmanConfigsDirPath(), envmanConfigFileName)
}

func ensureEnvmanConfigDirExists() error {
	confDirPth := getEnvmanConfigsDirPath()
	isExists, err := pathutil.IsDirExists(confDirPth)
	if !isExists || err != nil {
		if err := os.MkdirAll(confDirPth, 0777); err != nil {
			return err
		}
	}
	return nil
}

func createDefaultConfigsModel() ConfigsModel {
	return ConfigsModel{
		EnvBytesLimitInKB:     defaultEnvBytesLimitInKB,
		EnvListBytesLimitInKB: defaultEnvListBytesLimitInKB,
	}
}

// GetConfigs resolves the env var limits with default < configs.json < process
// env var precedence. A malformed or negative env override is ignored so a bad
// value never breaks a build; the caller setting it is expected to validate.
func GetConfigs() (ConfigsModel, error) {
	configPth := getEnvmanConfigsFilePath()
	configs := createDefaultConfigsModel()

	isExist, err := pathutil.IsPathExists(configPth)
	if err != nil {
		return ConfigsModel{}, err
	}

	if isExist {
		bytes, err := fileutil.ReadBytesFromFile(configPth)
		if err != nil {
			return ConfigsModel{}, err
		}

		type ConfigsFileMode struct {
			EnvBytesLimitInKB     *int `json:"env_bytes_limit_in_kb,omitempty"`
			EnvListBytesLimitInKB *int `json:"env_list_bytes_limit_in_kb,omitempty"`
		}

		var userConfigs ConfigsFileMode
		if err := json.Unmarshal(bytes, &userConfigs); err != nil {
			return ConfigsModel{}, err
		}

		if userConfigs.EnvBytesLimitInKB != nil {
			configs.EnvBytesLimitInKB = *userConfigs.EnvBytesLimitInKB
		}
		if userConfigs.EnvListBytesLimitInKB != nil {
			configs.EnvListBytesLimitInKB = *userConfigs.EnvListBytesLimitInKB
		}
	}

	if limit, ok := envLimitOverride(EnvBytesLimitInKBEnvKey); ok {
		configs.EnvBytesLimitInKB = limit
	}
	if limit, ok := envLimitOverride(EnvListBytesLimitInKBEnvKey); ok {
		configs.EnvListBytesLimitInKB = limit
	}

	return configs, nil
}

// envLimitOverride reads a limit from a process env var, treating a value of 0
// as "no limit" (matching validateEnv). Empty, non-integer, or negative values
// return ok=false so the file/default value stands.
func envLimitOverride(key string) (int, bool) {
	value := os.Getenv(key)
	if value == "" {
		return 0, false
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 0 {
		return 0, false
	}
	return limit, true
}

// saveConfigs ...
//
//	only used for unit testing at the moment
func saveConfigs(configModel ConfigsModel) error {
	if err := ensureEnvmanConfigDirExists(); err != nil {
		return err
	}

	bytes, err := json.Marshal(configModel)
	if err != nil {
		return err
	}
	configsPth := getEnvmanConfigsFilePath()
	return fileutil.WriteBytesToFile(configsPth, bytes)
}
