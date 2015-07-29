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
	Envs []models.EnvironmentItemModel `yaml:"envs"`
}

// -------------------
// --- Environment handling methods

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

func removeDefaults(env *models.EnvironmentItemModel) error {
	opts, err := env.GetOptions()
	if err != nil {
		return err
	}
	if opts.Title != nil && *opts.Title == "" {
		opts.Title = nil
	}
	if opts.Description != nil && *opts.Description == "" {
		opts.Description = nil
	}
	if opts.IsRequired != nil && *opts.IsRequired == models.DefaultIsRequired {
		opts.IsRequired = nil
	}
	if opts.IsExpand != nil && *opts.IsExpand == models.DefaultIsExpand {
		opts.IsExpand = nil
	}
	if opts.IsDontChangeValue != nil && *opts.IsDontChangeValue == models.DefaultIsDontChangeValue {
		opts.IsDontChangeValue = nil
	}
	(*env)[models.OptionsKey] = opts
	return nil
}

func generateFormattedYMLForEnvModels(envs []models.EnvironmentItemModel) ([]byte, error) {
	envMapSlice := []models.EnvironmentItemModel{}
	for _, env := range envs {
		err := removeDefaults(&env)
		if err != nil {
			return []byte{}, err
		}

		hasOptions := false
		opts, err := env.GetOptions()
		if err != nil {
			return []byte{}, err
		}

		if opts.Title != nil {
			hasOptions = true
		}
		if opts.Description != nil {
			hasOptions = true
		}
		if len(opts.ValueOptions) > 0 {
			hasOptions = true
		}
		if opts.IsRequired != nil {
			hasOptions = true
		}
		if opts.IsExpand != nil {
			hasOptions = true
		}
		if opts.IsDontChangeValue != nil {
			hasOptions = true
		}

		if !hasOptions {
			delete(env, models.OptionsKey)
		}

		envMapSlice = append(envMapSlice, env)
	}

	envYML := envsYMLModel{
		Envs: envMapSlice,
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
		if err := env.NormalizeEnvironmentItemModel(); err != nil {
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
