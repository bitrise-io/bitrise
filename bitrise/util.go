package bitrise

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-pathutil/pathutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// ExportEnvironmentsList ...
func ExportEnvironmentsList(envsList []envmanModels.EnvironmentItemModel) error {
	log.Debugln("[BITRISE_CLI] - Exporting environments:", envsList)

	for _, env := range envsList {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return err
		}

		if value != "" {
			if err := RunEnvmanAdd(key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

// CleanupStepWorkDir ...
func CleanupStepWorkDir() error {
	stepYMLPth := BitriseWorkDirPath + "/current_step.yml"
	if err := RemoveFile(stepYMLPth); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step yml: ", err))
	}

	stepDir := BitriseWorkStepsDirPath
	if err := RemoveDir(stepDir); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step work dir: ", err))
	}
	return nil
}

// SetBuildFailedEnv ...
func SetBuildFailedEnv(failed bool) error {
	statusStr := "0"
	if failed {
		statusStr = "1"
	}
	if err := os.Setenv("STEPLIB_BUILD_STATUS", statusStr); err != nil {
		return err
	}

	if err := os.Setenv("BITRISE_BUILD_STATUS", statusStr); err != nil {
		return err
	}
	return nil
}

// TimeToFormattedSeconds ...
func TimeToFormattedSeconds(t time.Duration, postfix string) string {
	sec := t.Seconds()
	if sec > 10.0 {
		return fmt.Sprintf("%.f%s", sec, postfix)
	} else if sec < 1.0 {
		return fmt.Sprintf("%.2f%s", sec, postfix)
	}
	return fmt.Sprintf("%.1f%s", sec, postfix)
}

// SaveConfigToFile ...
func SaveConfigToFile(pth string, bitriseConf models.BitriseDataModel) error {
	contBytes, err := generateYAML(bitriseConf)
	if err != nil {
		return err
	}
	if err := WriteBytesToFile(pth, contBytes); err != nil {
		return err
	}

	log.Println()
	log.Infoln("=> Init success!")
	log.Infoln("File created at path:", pth)

	return nil
}

func generateYAML(v interface{}) ([]byte, error) {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

// ReadBitriseConfig ...
func ReadBitriseConfig(pth string) (models.BitriseDataModel, error) {
	log.Debugln("-> ReadBitriseConfig")
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return models.BitriseDataModel{}, err
	} else if !isExists {
		return models.BitriseDataModel{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return models.BitriseDataModel{}, err
	}
	var bitriseData models.BitriseDataModel
	if err := yaml.Unmarshal(bytes, &bitriseData); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.Normalize(); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.Validate(); err != nil {
		return models.BitriseDataModel{}, err
	}

	if err := bitriseData.FillMissingDefaults(); err != nil {
		return models.BitriseDataModel{}, err
	}

	return bitriseData, nil
}

// ReadSpecStep ...
func ReadSpecStep(pth string) (stepmanModels.StepModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return stepmanModels.StepModel{}, err
	} else if !isExists {
		return stepmanModels.StepModel{}, errors.New(fmt.Sprint("No file found at path", pth))
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return stepmanModels.StepModel{}, err
	}

	var stepModel stepmanModels.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.Normalize(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.ValidateStep(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	return stepModel, nil
}

// WriteStringToFile ...
func WriteStringToFile(pth string, fileCont string) error {
	return WriteBytesToFile(pth, []byte(fileCont))
}

// WriteBytesToFile ...
func WriteBytesToFile(pth string, fileCont []byte) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorln("Failed to close file:", err)
		}
	}()

	if _, err := file.Write(fileCont); err != nil {
		return err
	}

	return nil
}

// IsVersionBetween ...
//  returns true if it's between the lower and upper limit
//  or in case it matches the lower or the upper limit
func IsVersionBetween(verBase, verLower, verUpper string) (bool, error) {
	r1, err := stepmanModels.CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	r2, err := stepmanModels.CompareVersions(verBase, verUpper)
	if err != nil {
		return false, err
	}
	if r2 == -1 {
		return false, nil
	}

	return true, nil
}

// IsVersionGreaterOrEqual ...
//  returns true if verBase is greater or equal to verLower
func IsVersionGreaterOrEqual(verBase, verLower string) (bool, error) {
	r1, err := stepmanModels.CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	return true, nil
}
