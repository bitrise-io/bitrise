package bitrise

import (
	"fmt"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
)

//=======================================
// Constants
//=======================================

const (
	bitriseVersionSetupStateFileName = "setup.version"
	bitriseConfigFileName            = "config.yml"
)

const (
	defaultOptOutAnalytics    = false
	defaultOptOutAnalyticsStr = "false"
)

//=======================================
// Models
//=======================================

// ConfigModel ...
type ConfigModel struct {
	OptOutAnalytics bool
}

// ConfigFileModel ...
type ConfigFileModel struct {
	OptOutAnalytics *string `yaml:"opt_out_analytics"`
}

// NewDefaultConfig ...
func NewDefaultConfig() ConfigModel {
	return ConfigModel{
		OptOutAnalytics: false,
	}
}

// NewConfigFromBytes ...
func NewConfigFromBytes(bytes []byte) (ConfigModel, error) {
	var fileConfig ConfigFileModel
	if err := yaml.Unmarshal(bytes, &fileConfig); err != nil {
		return ConfigModel{}, err
	}

	if fileConfig.OptOutAnalytics != nil && *fileConfig.OptOutAnalytics != "false" && *fileConfig.OptOutAnalytics != "true" {
		return ConfigModel{}, fmt.Errorf("Invalid config: opt_out_analytics value should be (\"false\" / \"true\"), actual: (%s)", *fileConfig.OptOutAnalytics)
	}

	config := fileConfig.convert()

	return config, nil
}

// Convert ConfigModel into ConfigFileModel
// Ommits every default value
func (c ConfigModel) convert() ConfigFileModel {
	config := ConfigFileModel{}

	if c.OptOutAnalytics != defaultOptOutAnalytics {
		config.OptOutAnalytics = pointers.NewStringPtr("true")
	}

	return config
}

// Convert ConfigFileModel into ConfigModel
// Override every ConfigModel default value, with ConfigFileModel values
func (c ConfigFileModel) convert() ConfigModel {
	config := NewDefaultConfig()

	if c.OptOutAnalytics != nil {
		if *c.OptOutAnalytics == "true" {
			config.OptOutAnalytics = true
		} else {
			config.OptOutAnalytics = false
		}
	}

	return config
}

//=======================================
// Utility
//=======================================

func getBitriseConfigVersionSetupFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), bitriseVersionSetupStateFileName)
}

//=======================================
// Main
//=======================================

// GetBitriseConfigFilePath ...
func GetBitriseConfigFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), bitriseConfigFileName)
}

// GetBitriseConfigsDirPath ...
func GetBitriseConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".bitrise")
}

// EnsureBitriseConfigDirExists ...
func EnsureBitriseConfigDirExists() error {
	confDirPth := GetBitriseConfigsDirPath()
	return pathutil.EnsureDirExist(confDirPth)
}

// CheckIsSetupWasDoneForVersion ...
func CheckIsSetupWasDoneForVersion(ver string) bool {
	configPth := getBitriseConfigVersionSetupFilePath()
	cont, err := fileutil.ReadStringFromFile(configPth)
	if err != nil {
		return false
	}
	return (cont == ver)
}

// SaveSetupSuccessForVersion ...
func SaveSetupSuccessForVersion(ver string) error {
	if err := EnsureBitriseConfigDirExists(); err != nil {
		return err
	}
	configPth := getBitriseConfigVersionSetupFilePath()
	return fileutil.WriteStringToFile(configPth, ver)
}

// ReadConfig ...
func ReadConfig() (ConfigModel, error) {
	config := NewDefaultConfig()

	configPth := GetBitriseConfigFilePath()
	if exist, err := pathutil.IsPathExists(configPth); err != nil {
		return ConfigModel{}, err
	} else if exist {
		bytes, err := fileutil.ReadBytesFromFile(configPth)
		if err != nil {
			return ConfigModel{}, err
		}

		if config, err = NewConfigFromBytes(bytes); err != nil {
			return ConfigModel{}, err
		}
	}

	return config, nil
}

// SaveConfig ...
func SaveConfig(config ConfigModel) error {
	// Converte config to file config
	fileConfig := config.convert()

	bytes, err := yaml.Marshal(fileConfig)
	if err != nil {
		return err
	}

	configPth := GetBitriseConfigFilePath()
	return fileutil.WriteBytesToFile(configPth, bytes)
}
