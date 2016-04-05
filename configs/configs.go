package configs

import (
	"fmt"
	"path"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/codegangsta/cli"
)

// ---------------------------
// --- Project level vars / configs

var (
	// IsCIMode ...
	IsCIMode = false
	// IsDebugMode ...
	IsDebugMode = false
	// IsPullRequestMode ...
	IsPullRequestMode = false
)

// OutputFormat ...
var OutputFormat = OutputFormatRaw

// ---------------------------
// --- Consts

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
)

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

const (
	bitriseVersionSetupStateFileName = "setup.version"
)

// GetBitriseConfigsDirPath ...
func GetBitriseConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".bitrise")
}

func getBitriseConfigVersionSetupFilePath() string {
	return path.Join(GetBitriseConfigsDirPath(), bitriseVersionSetupStateFileName)
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
