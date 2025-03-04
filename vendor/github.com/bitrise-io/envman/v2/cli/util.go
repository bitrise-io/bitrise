package cli

import (
	"errors"

	"github.com/bitrise-io/envman/v2/env"
	"github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"gopkg.in/yaml.v2"
)

var (
	// CurrentEnvStoreFilePath ...
	CurrentEnvStoreFilePath string

	// ToolMode ...
	ToolMode bool
)

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
	if opts.Category != nil && *opts.Category == "" {
		opts.Category = nil
	}
	if opts.Summary != nil && *opts.Summary == "" {
		opts.Summary = nil
	}
	if opts.IsRequired != nil && *opts.IsRequired == models.DefaultIsRequired {
		opts.IsRequired = nil
	}
	if opts.IsDontChangeValue != nil && *opts.IsDontChangeValue == models.DefaultIsDontChangeValue {
		opts.IsDontChangeValue = nil
	}
	if opts.IsTemplate != nil && *opts.IsTemplate == models.DefaultIsTemplate {
		opts.IsTemplate = nil
	}
	if opts.IsExpand != nil && *opts.IsExpand == models.DefaultIsExpand {
		opts.IsExpand = nil
	}
	if opts.IsSensitive != nil && *opts.IsSensitive == models.DefaultIsSensitive {
		opts.IsSensitive = nil
	}
	if opts.SkipIfEmpty != nil && *opts.SkipIfEmpty == models.DefaultSkipIfEmpty {
		opts.SkipIfEmpty = nil
	}
	if opts.Unset != nil && *opts.Unset == models.DefaultUnset {
		opts.Unset = nil
	}

	(*env)[models.OptionsKey] = opts
	return nil
}

func generateFormattedYMLForEnvModels(envs []models.EnvironmentItemModel) (models.EnvsSerializeModel, error) {
	envMapSlice := []models.EnvironmentItemModel{}
	for _, env := range envs {
		err := removeDefaults(&env)
		if err != nil {
			return models.EnvsSerializeModel{}, err
		}

		hasOptions := false
		opts, err := env.GetOptions()
		if err != nil {
			return models.EnvsSerializeModel{}, err
		}

		if opts.Title != nil {
			hasOptions = true
		}
		if opts.Description != nil {
			hasOptions = true
		}
		if opts.Summary != nil {
			hasOptions = true
		}
		if len(opts.ValueOptions) > 0 {
			hasOptions = true
		}
		if opts.IsRequired != nil {
			hasOptions = true
		}
		if opts.IsDontChangeValue != nil {
			hasOptions = true
		}
		if opts.IsTemplate != nil {
			hasOptions = true
		}
		if opts.IsExpand != nil {
			hasOptions = true
		}
		if opts.IsSensitive != nil {
			hasOptions = true
		}
		if opts.SkipIfEmpty != nil {
			hasOptions = true
		}
		if opts.Unset != nil {
			hasOptions = true
		}

		if !hasOptions {
			delete(env, models.OptionsKey)
		}

		envMapSlice = append(envMapSlice, env)
	}

	return models.EnvsSerializeModel{
		Envs: envMapSlice,
	}, nil
}

// -------------------
// --- File methods

// WriteEnvMapToFile ...
func WriteEnvMapToFile(pth string, envs []models.EnvironmentItemModel) error {
	if pth == "" {
		return errors.New("No path provided")
	}

	envYML, err := generateFormattedYMLForEnvModels(envs)
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(envYML)
	if err != nil {
		return err
	}
	return fileutil.WriteBytesToFile(pth, bytes)
}

func initAtPath(pth string) error {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return err
	} else if !exist {
		if err := WriteEnvMapToFile(pth, []models.EnvironmentItemModel{}); err != nil {
			return err
		}
	} else {
		return errors.New("Path already exist: " + pth)
	}
	return nil
}

// ParseEnvsYML ...
func ParseEnvsYML(bytes []byte) ([]models.EnvironmentItemModel, error) {
	var envsYML models.EnvsSerializeModel
	if err := yaml.Unmarshal(bytes, &envsYML); err != nil {
		return []models.EnvironmentItemModel{}, err
	}
	for _, env := range envsYML.Envs {
		if err := env.NormalizeValidateFillDefaults(); err != nil {
			return []models.EnvironmentItemModel{}, err
		}
	}
	return envsYML.Envs, nil
}

// ReadEnvs ...
func ReadEnvs(pth string) ([]models.EnvironmentItemModel, error) {
	bytes, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return []models.EnvironmentItemModel{}, err
	}

	return ParseEnvsYML(bytes)
}

func evaluateEnvs(newEnvs []models.EnvironmentItemModel, envSource env.EnvironmentSource) ([]string, error) {
	result, err := env.GetDeclarationsSideEffects(newEnvs, envSource)
	if err != nil {
		return nil, err
	}
	var envs []string
	for key, value := range result.ResultEnvironment {
		envs = append(envs, key+"="+value)
	}
	return envs, nil
}

// ReadAndEvaluateEnvs ...
func ReadAndEvaluateEnvs(envStorePth string, envSource env.EnvironmentSource) ([]string, error) {
	envs, err := ReadEnvs(envStorePth)
	if err != nil {
		return nil, err
	}
	return evaluateEnvs(envs, envSource)
}

// ReadEnvsOrCreateEmptyList ...
func ReadEnvsOrCreateEmptyList(envStorePth string) ([]models.EnvironmentItemModel, error) {
	envModels, err := ReadEnvs(envStorePth)
	if err != nil {
		if err.Error() == "No environment variable list found" {
			err = initAtPath(envStorePth)
			return []models.EnvironmentItemModel{}, err
		}
		return []models.EnvironmentItemModel{}, err
	}
	return envModels, nil
}
