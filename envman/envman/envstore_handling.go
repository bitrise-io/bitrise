package envman

import (
	"errors"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil"
	"gopkg.in/yaml.v2"
)

var (
	CurrentEnvStoreFilePath string
)

func InitAtPath(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if exist == false {
		if err := WriteEnvMapToFile(pth, []EnvModel{}); err != nil {
			return err
		}
	}
	return nil
}

func LoadEnvMap() ([]EnvModel, error) {
	if envsYML, err := readEnvMapFromFile(CurrentEnvStoreFilePath); err != nil {
		return []EnvModel{}, err
	} else {
		return envsYML.convertToEnvModelArray(), nil
	}
}

func LoadEnvMapOrCreate() ([]EnvModel, error) {
	if envModels, err := LoadEnvMap(); err != nil {
		if err.Error() == "No environment variable list found" {
			err = InitAtPath(CurrentEnvStoreFilePath)
			return []EnvModel{}, err
		}
		return []EnvModel{}, err
	} else {
		return envModels, nil
	}
}

func UpdateOrAddToEnvlist(envs []EnvModel, env EnvModel) ([]EnvModel, error) {
	var newEnvs []EnvModel
	exist := false

	for _, eModel := range envs {
		if eModel.Key == env.Key {
			exist = true
			newEnvs = append(newEnvs, env)
		} else {
			newEnvs = append(newEnvs, eModel)
		}
	}

	if exist == false {
		newEnvs = append(newEnvs, env)
	}

	if err := WriteEnvMapToFile(CurrentEnvStoreFilePath, newEnvs); err != nil {
		return []EnvModel{}, err
	}

	return newEnvs, nil
}

func readEnvMapFromFile(pth string) (envsYMLModel, error) {

	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return envsYMLModel{}, err
	} else if isExists == false {
		return envsYMLModel{}, errors.New("No environment variable list found")
	}

	if bytes, err := ioutil.ReadFile(pth); err != nil {
		return envsYMLModel{}, err
	} else {
		var envsModel envsYMLModel
		if err := yaml.Unmarshal(bytes, &envsModel); err != nil {
			return envsYMLModel{}, err
		}

		return envsModel, nil
	}
}

func generateFormattedYMLForEnvModels(envs []EnvModel) ([]byte, error) {
	envYML := convertToEnvsYMLModel(envs)
	if bytes, err := yaml.Marshal(envYML); err != nil {
		return []byte{}, err
	} else {
		return bytes, nil
	}
}

func WriteEnvMapToFile(pth string, envs []EnvModel) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	if file, err := os.Create(pth); err != nil {
		return err
	} else {
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatalln("Failed to close file:", err)
			}
		}()

		if jsonContBytes, err := generateFormattedYMLForEnvModels(envs); err != nil {
			return err
		} else if _, err := file.Write(jsonContBytes); err != nil {
			return err
		}
		return nil
	}
}
