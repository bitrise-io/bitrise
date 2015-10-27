package envman

import (
	"encoding/json"
	"os"
	"path"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	envmanConfigFileName         = "configs.json"
	defaultEnvBytesLimitInKB     = 20
	defaultEnvListBytesLimitInKB = 100
)

// Configs ...
type Configs struct {
	EnvBytesLimitInKB     int `json:"env_bytes_limit_in_kb,omitempty" yaml:"env_bytes_limit_in_kb,omitempty"`
	EnvListBytesLimitInKB int `json:"env_list_bytes_limit_in_kb,omitempty" yaml:"env_list_bytes_limit_in_kb,omitempty"`
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

// CheckIfConfigsSaved ...
func CheckIfConfigsSaved() bool {
	configPth := getEnvmanConfigsFilePath()
	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return false
	}

	var configs Configs
	if err := json.Unmarshal(bytes, &configs); err != nil {
		return false
	}
	return true
}

// ReadConfigs ...
func ReadConfigs() (Configs, error) {
	configPth := getEnvmanConfigsFilePath()
	bytes, err := fileutil.ReadBytesFromFile(configPth)
	if err != nil {
		return Configs{}, err
	}

	var configs Configs
	if err := json.Unmarshal(bytes, &configs); err != nil {
		return Configs{}, err
	}
	return configs, nil
}

// SaveDefaultConfigs ...
func SaveDefaultConfigs() error {
	if err := ensureEnvmanConfigDirExists(); err != nil {
		return err
	}

	defaultConfigs := Configs{
		EnvBytesLimitInKB:     defaultEnvBytesLimitInKB,
		EnvListBytesLimitInKB: defaultEnvListBytesLimitInKB,
	}
	bytes, err := json.Marshal(defaultConfigs)
	if err != nil {
		return err
	}
	configsPth := getEnvmanConfigsFilePath()
	return fileutil.WriteBytesToFile(configsPth, bytes)
}
