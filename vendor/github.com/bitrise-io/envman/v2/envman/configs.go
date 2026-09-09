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

// ConfigLimitOverrides selects which env size limits to write to the config
// file. A nil field is left unchanged; a non-nil field is written as-is,
// including 0, which disables that check.
type ConfigLimitOverrides struct {
	EnvBytesLimitInKB     *int
	EnvListBytesLimitInKB *int
}

// configsFile mirrors the on-disk config, using pointers so an unset field is
// omitted and a 0 is written explicitly (unlike ConfigsModel, whose omitempty
// int fields would drop a 0 and fall back to the default).
type configsFile struct {
	EnvBytesLimitInKB     *int `json:"env_bytes_limit_in_kb,omitempty"`
	EnvListBytesLimitInKB *int `json:"env_list_bytes_limit_in_kb,omitempty"`
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

		var userConfigs configsFile
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

// SetConfigLimits merges the given limits into the config file, creating it if
// needed, and returns a restore function that reverts the file to its previous
// state (its original content, or its absence).
//
// Unlike the process env var overrides, this reaches an `envman` binary that
// predates env var support (e.g. one invoked by a step to export outputs),
// since reading configs.json is the older, always-present mechanism.
func SetConfigLimits(overrides ConfigLimitOverrides) (restore func() error, err error) {
	if overrides.EnvBytesLimitInKB == nil && overrides.EnvListBytesLimitInKB == nil {
		return func() error { return nil }, nil
	}

	configPth := getEnvmanConfigsFilePath()

	existed, err := pathutil.IsPathExists(configPth)
	if err != nil {
		return nil, err
	}

	var merged configsFile
	var originalBytes []byte
	if existed {
		originalBytes, err = fileutil.ReadBytesFromFile(configPth)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(originalBytes, &merged); err != nil {
			return nil, err
		}
	}

	if overrides.EnvBytesLimitInKB != nil {
		merged.EnvBytesLimitInKB = overrides.EnvBytesLimitInKB
	}
	if overrides.EnvListBytesLimitInKB != nil {
		merged.EnvListBytesLimitInKB = overrides.EnvListBytesLimitInKB
	}

	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}

	if err := ensureEnvmanConfigDirExists(); err != nil {
		return nil, err
	}
	if err := fileutil.WriteBytesToFile(configPth, mergedBytes); err != nil {
		return nil, err
	}

	return func() error {
		if !existed {
			if err := os.Remove(configPth); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		return fileutil.WriteBytesToFile(configPth, originalBytes)
	}, nil
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
