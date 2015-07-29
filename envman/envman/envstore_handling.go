package envman

import (
	"errors"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"gopkg.in/yaml.v2"
)

var (
	// CurrentEnvStoreFilePath ...
	CurrentEnvStoreFilePath string

	// ToolMode ...
	ToolMode bool
)

// ClearPathIfExist ...
func ClearPathIfExist(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(pth); err != nil {
			return err
		}
	}
	return nil
}

// InitAtPath ...
func InitAtPath(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exist {
		if err := WriteEnvMapToFile(pth, []models.EnvironmentItemModel{}); err != nil {
			return err
		}
	} else {
		errorMsg := "Path already exist: " + pth
		return errors.New(errorMsg)
	}
	return nil
}

func readEnvMapFromFile(pth string) (envsYMLModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return envsYMLModel{}, err
	} else if !isExists {
		return envsYMLModel{}, errors.New("No environment variable list found")
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return envsYMLModel{}, err
	}
	var envsModel envsYMLModel
	if err := yaml.Unmarshal(bytes, &envsModel); err != nil {
		return envsYMLModel{}, err
	}

	return envsModel, nil
}

// LoadEnvMap ...
func LoadEnvMap() ([]models.EnvironmentItemModel, error) {
	envsYML, err := readEnvMapFromFile(CurrentEnvStoreFilePath)
	if err != nil {
		return []models.EnvironmentItemModel{}, err
	}
	return envsYML.Envs, nil
}

// LoadEnvMapOrCreate ...
func LoadEnvMapOrCreate() ([]models.EnvironmentItemModel, error) {
	envModels, err := LoadEnvMap()
	if err != nil {
		if err.Error() == "No environment variable list found" {
			err = InitAtPath(CurrentEnvStoreFilePath)
			return []models.EnvironmentItemModel{}, err
		}
		return []models.EnvironmentItemModel{}, err
	}
	return envModels, nil
}

// UpdateOrAddToEnvlist ...
func UpdateOrAddToEnvlist(oldEnvSlice []models.EnvironmentItemModel, newEnv models.EnvironmentItemModel, replace bool) ([]models.EnvironmentItemModel, error) {
	newKey, _, err := newEnv.GetKeyValuePair()
	if err != nil {
		return []models.EnvironmentItemModel{}, err
	}

	var newEnvs []models.EnvironmentItemModel
	exist := false

	if replace {
		match := 0
		for _, env := range oldEnvSlice {
			key, _, err := env.GetKeyValuePair()
			if err != nil {
				return []models.EnvironmentItemModel{}, err
			}

			if key == newKey {
				match = match + 1
			}
		}
		if match > 1 {
			if ToolMode {
				return []models.EnvironmentItemModel{}, errors.New("More then one env exist with key '" + newKey + "'")
			}
			msg := "   More then one env exist with key '" + newKey + "' replace all/append ['replace/append'] ?"
			answer, err := goinp.AskForString(msg)
			if err != nil {
				return []models.EnvironmentItemModel{}, err
			}

			switch answer {
			case "replace":
				break
			case "append":
				replace = false
				break
			default:
				return []models.EnvironmentItemModel{}, errors.New("Failed to parse answer: '" + answer + "' use ['replace/append']!")
			}
		}
	}

	for _, env := range oldEnvSlice {
		key, _, err := env.GetKeyValuePair()
		if err != nil {
			return []models.EnvironmentItemModel{}, err
		}

		if replace && key == newKey {
			exist = true
			newEnvs = append(newEnvs, newEnv)
		} else {
			newEnvs = append(newEnvs, env)
		}
	}

	if !exist {
		newEnvs = append(newEnvs, newEnv)
	}

	if err := WriteEnvMapToFile(CurrentEnvStoreFilePath, newEnvs); err != nil {
		return []models.EnvironmentItemModel{}, err
	}
	return newEnvs, nil
}

func generateFormattedYMLForEnvModels(envs []models.EnvironmentItemModel) ([]byte, error) {
	envYML := envsYMLModel{
		Envs: envs,
	}
	bytes, err := yaml.Marshal(envYML)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

// WriteEnvMapToFile ...
func WriteEnvMapToFile(pth string, envs []models.EnvironmentItemModel) error {
	log.Info("Write")
	if pth == "" {
		return errors.New("No path provided")
	}

	file, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln("[ENVMAN] - Failed to close file:", err)
		}
	}()

	if jsonContBytes, err := generateFormattedYMLForEnvModels(envs); err != nil {
		return err
	} else if _, err := file.Write(jsonContBytes); err != nil {
		return err
	}
	return nil
}
