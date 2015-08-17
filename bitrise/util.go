package bitrise

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	models "github.com/bitrise-io/bitrise/models/models_1_0_0"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
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
	if err := cmdex.RemoveFile(stepYMLPth); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step yml: ", err))
	}

	stepDir := BitriseWorkStepsDirPath
	if err := cmdex.RemoveDir(stepDir); err != nil {
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
	if err := fileutil.WriteBytesToFile(pth, contBytes); err != nil {
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

func normalizeValidateFillMissingDefaults(bitriseData *models.BitriseDataModel) error {
	if err := bitriseData.Normalize(); err != nil {
		return err
	}
	if err := bitriseData.Validate(); err != nil {
		return err
	}
	if err := bitriseData.FillMissingDefaults(); err != nil {
		return err
	}
	return nil
}

// ConfigModelFromYAMLBytes ...
func ConfigModelFromYAMLBytes(configBytes []byte) (bitriseData models.BitriseDataModel, err error) {
	if err = yaml.Unmarshal(configBytes, &bitriseData); err != nil {
		return
	}

	if err = normalizeValidateFillMissingDefaults(&bitriseData); err != nil {
		return
	}

	return
}

// ConfigModelFromJSONBytes ...
func ConfigModelFromJSONBytes(configBytes []byte) (bitriseData models.BitriseDataModel, err error) {
	if err = json.Unmarshal(configBytes, &bitriseData); err != nil {
		return
	}

	if err = normalizeValidateFillMissingDefaults(&bitriseData); err != nil {
		return
	}

	return
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

	if strings.HasSuffix(pth, ".json") {
		log.Debugln("=> Using JSON parser for: ", pth)
		return ConfigModelFromJSONBytes(bytes)
	}

	log.Debugln("=> Using YAML parser for: ", pth)
	return ConfigModelFromYAMLBytes(bytes)
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

	if err := stepModel.ValidateStep(false); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if err := stepModel.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	return stepModel, nil
}
