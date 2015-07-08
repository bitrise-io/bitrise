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
	// CurrentEnvStoreFilePath ...
	CurrentEnvStoreFilePath string
)

// InitAtPath ...
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

// LoadEnvMap ...
func LoadEnvMap() ([]EnvModel, error) {
	envsYML, err := readEnvMapFromFile(CurrentEnvStoreFilePath)
	if err != nil {
		return []EnvModel{}, err
	}
	return envsYML.convertToEnvModelArray(), nil
}

// LoadEnvMapOrCreate ...
func LoadEnvMapOrCreate() ([]EnvModel, error) {
	envModels, err := LoadEnvMap()
	if err != nil {
		if err.Error() == "No environment variable list found" {
			err = InitAtPath(CurrentEnvStoreFilePath)
			return []EnvModel{}, err
		}
		return []EnvModel{}, err
	}
	return envModels, nil
}

// UpdateOrAddToEnvlist ...
func UpdateOrAddToEnvlist(envs []EnvModel, env EnvModel) ([]EnvModel, error) {
	newEnvs := append(envs, env)
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

func generateFormattedYMLForEnvModels(envs []EnvModel) ([]byte, error) {
	envYML := convertToEnvsYMLModel(envs)
	bytes, err := yaml.Marshal(envYML)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

// WriteEnvMapToFile ...
func WriteEnvMapToFile(pth string, envs []EnvModel) error {
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
