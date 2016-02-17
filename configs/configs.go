package configs

import (
	"fmt"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/codegangsta/cli"
)

//=======================================
// Consts
//=======================================

const (
	// CIModeEnvKey ...
	CIModeEnvKey = "CI"
	// PRModeEnvKey ...
	PRModeEnvKey = "PR"
	// PullRequestIDEnvKey ...
	PullRequestIDEnvKey = "PULL_REQUEST_ID"
	// DebugModeEnvKey ...
	DebugModeEnvKey = "DEBUG"
	// LogLevelEnvKey ...
	LogLevelEnvKey = "LOGLEVEL"
	// IsAnalyticsDisabledEnvKey ...
	IsAnalyticsDisabledEnvKey = "IS_ANALYTICS_DISABLED"
)

const (
	fileNameBitriseVersionSetupState = "setup.version"
	fileNameBitriseConfig            = "config.yml"
)

const (
	// OuputFormatKey ...
	OuputFormatKey = "format"
	// OutputFormatRaw ...
	OutputFormatRaw = "raw"
	// OutputFormatJSON ...
	OutputFormatJSON = "json"
	// OutputFormatYML ...
	OutputFormatYML = "yml"
)

//=======================================
// Project level vars / configs
//=======================================

var (
	// IsCIMode ...
	IsCIMode = false
	// IsDebugMode ...
	IsDebugMode = false
	// IsPullRequestMode ...
	IsPullRequestMode = false
	// IsAnalyticsDisabled ...
	IsAnalyticsDisabled = false
	// OutputFormat ...
	OutputFormat = OutputFormatRaw
)

//=======================================
// Models
//=======================================

// ConfigModel ...
type ConfigModel struct {
	IsAnalyticsDisabled bool `yaml:"is_analytics_disabled"`
}

// NewConfigFromBytes ...
func NewConfigFromBytes(bytes []byte) (ConfigModel, error) {
	var config ConfigModel
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return ConfigModel{}, err
	}

	return config, nil
}

//=======================================
// Utility
//=======================================

func getBitriseConfigVersionSetupFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), fileNameBitriseVersionSetupState)
}

//=======================================
// Main
//=======================================

// ConfigureOutputFormat ...
func ConfigureOutputFormat(c *cli.Context) error {
	outFmt := c.String(OuputFormatKey)
	switch outFmt {
	case OutputFormatRaw, OutputFormatJSON, OutputFormatYML:
		// valid
		OutputFormat = outFmt
	case "":
		// default
		OutputFormat = OutputFormatRaw
	default:
		// invalid
		return fmt.Errorf("Invalid Output Format: %s", outFmt)
	}
	return nil
}

// GetBitriseConfigFilePath ...
func GetBitriseConfigFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), fileNameBitriseConfig)
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
	config := ConfigModel{}

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
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	configPth := GetBitriseConfigFilePath()
	return fileutil.WriteBytesToFile(configPth, bytes)
}
