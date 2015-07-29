package envman

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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

type envsYMLModel struct {
	Envs []models.EnvironmentItemModel `yaml:"environments"`
}

// -------------------
// --- Environment handling methods

func prepareRawEnv(rawEnv *models.EnvironmentItemModel) error {
	if err := rawEnv.Normalize(); err != nil {
		return err
	}

	if err := validate(*rawEnv); err != nil {
		return err
	}

	if err := rawEnv.FillMissingDeafults(); err != nil {
		return err
	}
	return nil
}

// Validate ...
func validate(env models.EnvironmentItemModel) error {
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("Invalid environment: empty env_key")
	}
	_, err = env.GetOptions()
	if err != nil {
		return err
	}
	return nil
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
	envMapSlice := []map[string]interface{}{}
	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return []byte{}, err
		}

		hasOptions := false
		opts, err := env.GetOptions()
		if err != nil {
			return []byte{}, err
		}

		envOptionsMap := map[string]interface{}{}
		if opts.Title != nil && *opts.Title != "" {
			envOptionsMap["title"] = *opts.Title
			hasOptions = true
		}
		if *opts.Description != "" {
			envOptionsMap["description"] = *opts.Description
			hasOptions = true
		}
		if len(opts.ValueOptions) > 0 {
			envOptionsMap["value_options"] = opts.ValueOptions
			hasOptions = true
		}
		if *opts.IsRequired != models.DefaultIsRequired {
			envOptionsMap["is_required"] = *opts.IsRequired
			hasOptions = true
		}
		if *opts.IsExpand != models.DefaultIsExpand {
			envOptionsMap["is_expand"] = *opts.IsExpand
			hasOptions = true
		}
		if *opts.IsDontChangeValue != models.DefaultIsDontChangeValue {
			envOptionsMap["is_dont_change_value"] = *opts.IsDontChangeValue
			hasOptions = true
		}

		envMap := map[string]interface{}{
			key: value,
		}
		if hasOptions {
			envMap[models.OptionsKey] = envOptionsMap
		}

		envMapSlice = append(envMapSlice, envMap)
	}
	envYML := map[string]interface{}{
		"environments": envMapSlice,
	}
	bytes, err := yaml.Marshal(envYML)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

// -------------------
// --- File methods

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

// ReadEnvs ...
func ReadEnvs(pth string) ([]models.EnvironmentItemModel, error) {
	if isExists, err := pathutil.IsPathExists(pth); err != nil {
		return []models.EnvironmentItemModel{}, err
	} else if !isExists {
		return []models.EnvironmentItemModel{}, errors.New("No environment variable list found")
	}

	bytes, err := ioutil.ReadFile(pth)
	if err != nil {
		return []models.EnvironmentItemModel{}, err
	}
	var envsYML envsYMLModel
	if err := yaml.Unmarshal(bytes, &envsYML); err != nil {
		return []models.EnvironmentItemModel{}, err
	}

	for _, env := range envsYML.Envs {
		if err := prepareRawEnv(&env); err != nil {
			return []models.EnvironmentItemModel{}, err
		}
	}
	return envsYML.Envs, nil
}

// ReadEnvsOrCreateEmptyList ...
func ReadEnvsOrCreateEmptyList() ([]models.EnvironmentItemModel, error) {
	envModels, err := ReadEnvs(CurrentEnvStoreFilePath)
	if err != nil {
		if err.Error() == "No environment variable list found" {
			err = InitAtPath(CurrentEnvStoreFilePath)
			return []models.EnvironmentItemModel{}, err
		}
		return []models.EnvironmentItemModel{}, err
	}
	return envModels, nil
}

// WriteEnvMapToFile ...
func WriteEnvMapToFile(pth string, envs []models.EnvironmentItemModel) error {
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

// -------------------
// --- Common methods

// ParseBool ...
func ParseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}

	lowercased := strings.ToLower(s)
	if lowercased == "yes" || lowercased == "y" {
		return true
	}
	if lowercased == "no" || lowercased == "n" {
		return false
	}

	value, err := strconv.ParseBool(s)
	if err != nil {
		log.Errorln("[ENVMAN] - isExpand: Failed to parse input:", err)
		return defaultValue
	}
	return value
}
