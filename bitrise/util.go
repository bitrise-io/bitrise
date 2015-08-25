package bitrise

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
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

// CollectEnvironmentsFromFile ...
func CollectEnvironmentsFromFile(pth string) ([]envmanModels.EnvironmentItemModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	var envstore envmanModels.EnvsYMLModel
	if err := yaml.Unmarshal(bytes, &envstore); err != nil {
		return []envmanModels.EnvironmentItemModel{}, err
	}

	for _, env := range envstore.Envs {
		if err := env.Normalize(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
		if err := env.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
		if err := env.Validate(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, err
		}
	}

	return envstore.Envs, nil
}

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
			if err := EnvmanAdd(InputEnvstorePath, key, value, *opts.IsExpand); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}
	return nil
}

// CleanupStepWorkDir ...
func CleanupStepWorkDir() error {
	stepYMLPth := path.Join(BitriseWorkDirPath, "current_step.yml")
	if err := cmdex.RemoveFile(stepYMLPth); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step yml: ", err))
	}

	stepDir := BitriseWorkStepsDirPath
	if err := cmdex.RemoveDir(stepDir); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step work dir: ", err))
	}
	return nil
}

// GetBuildFailedEnvironments ...
func GetBuildFailedEnvironments(failed bool) []string {
	statusStr := "0"
	if failed {
		statusStr = "1"
	}

	environments := []string{}
	steplibBuildStatusEnv := "STEPLIB_BUILD_STATUS" + "=" + statusStr
	environments = append(environments, steplibBuildStatusEnv)

	bitriseBuildStatusEnv := "BITRISE_BUILD_STATUS" + "=" + statusStr
	environments = append(environments, bitriseBuildStatusEnv)
	return environments
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

	bytes, err := fileutil.ReadBytesFromFile(pth)
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

	bytes, err := fileutil.ReadBytesFromFile(pth)
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

func fillStepOutputs(stepListItem *models.StepListItemModel, defaultStepLibSource string) error {
	// Create stepIDData
	compositeStepIDStr, workflowStep, err := models.GetStepIDStepDataPair(*stepListItem)
	if err != nil {
		return err
	}
	stepIDData, err := models.CreateStepIDDataFromString(compositeStepIDStr, defaultStepLibSource)
	if err != nil {
		return err
	}

	// Activate step - get step.yml
	tempStepCloneDirPath, err := pathutil.NormalizedOSTempDirPath("step_clone")
	if err != nil {
		return err
	}
	tempStepYMLDirPath, err := pathutil.NormalizedOSTempDirPath("step_yml")
	if err != nil {
		return err
	}
	tempStepYMLFilePath := path.Join(tempStepYMLDirPath, "step.yml")

	if stepIDData.SteplibSource == "path" {
		stepAbsLocalPth, err := pathutil.AbsPath(stepIDData.IDorURI)
		if err != nil {
			return err
		}
		if err := cmdex.CopyFile(path.Join(stepAbsLocalPth, "step.yml"), tempStepYMLFilePath); err != nil {
			return err
		}
	} else if stepIDData.SteplibSource == "git" {
		if err := cmdex.GitCloneTagOrBranch(stepIDData.IDorURI, tempStepCloneDirPath, stepIDData.Version); err != nil {
			return err
		}
		if err := cmdex.CopyFile(path.Join(tempStepCloneDirPath, "step.yml"), tempStepYMLFilePath); err != nil {
			return err
		}
	} else if stepIDData.SteplibSource == "_" {
		// Steplib independent steps are completly defined in workflow
		tempStepYMLFilePath = ""
	} else if stepIDData.SteplibSource != "" {
		if err := StepmanSetup(stepIDData.SteplibSource); err != nil {
			return err
		}
		if err := StepmanActivate(stepIDData.SteplibSource, stepIDData.IDorURI, stepIDData.Version, tempStepCloneDirPath, tempStepYMLFilePath); err != nil {
			return err
		}
	} else {
		return errors.New("Failed to fill step ouputs: unkown SteplibSource")
	}

	// Fill outputs
	if tempStepYMLFilePath != "" {
		specStep, err := ReadSpecStep(tempStepYMLFilePath)
		if err != nil {
			return err
		}

		workflowStep.Outputs = specStep.Outputs
		(*stepListItem)[compositeStepIDStr] = workflowStep
	}

	// Cleanup
	if err := cmdex.RemoveDir(tempStepCloneDirPath); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step clone dir: ", err))
	}
	if err := cmdex.RemoveDir(tempStepYMLDirPath); err != nil {
		return errors.New(fmt.Sprint("Failed to remove step clone dir: ", err))
	}

	return nil
}

// RemoveConfigRedundantFieldsAndFillStepOutputs ...
func RemoveConfigRedundantFieldsAndFillStepOutputs(config models.BitriseDataModel) error {
	for _, workflow := range config.Workflows {
		for _, stepListItem := range workflow.Steps {
			if err := fillStepOutputs(&stepListItem, config.DefaultStepLibSource); err != nil {
				return err
			}
		}
	}
	if err := config.RemoveRedundantFields(); err != nil {
		return err
	}
	return nil
}
