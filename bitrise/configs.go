package bitrise

import (
	"os"
	"path"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

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

// EnsureDir ...
func EnsureDir(dir string) error {
	isExists, err := pathutil.IsDirExists(dir)
	if !isExists || err != nil {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

// EnsureBitriseConfigDirExists ...
func EnsureBitriseConfigDirExists() error {
	confDirPth := GetBitriseConfigsDirPath()
	return EnsureDir(confDirPth)
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
