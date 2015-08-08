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

func getBitriseConfigsDirPath() string {
	return path.Join(pathutil.UserHomeDir(), ".bitrise")
}

func getBitriseConfigVersionSetupFilePath() string {
	return path.Join(getBitriseConfigsDirPath(), bitriseVersionSetupStateFileName)
}

func ensureBitriseConfigDirExists() error {
	confDirPth := getBitriseConfigsDirPath()
	isExists, err := pathutil.IsDirExists(confDirPth)
	if !isExists || err != nil {
		if err := os.MkdirAll(confDirPth, 0777); err != nil {
			return err
		}
	}
	return nil
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
	if err := ensureBitriseConfigDirExists(); err != nil {
		return err
	}
	configPth := getBitriseConfigVersionSetupFilePath()
	return fileutil.WriteStringToFile(configPth, ver)
}
