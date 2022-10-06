package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var errTimeout = errors.New("timeout")

func add(c *cli.Context) error {
	log.Debugln("[ENVMAN] Work path:", CurrentEnvStoreFilePath)

	key := c.String(KeyKey)
	expand := !c.Bool(NoExpandKey)
	replace := !c.Bool(AppendKey)
	skipIfEmpty := c.Bool(SkipIfEmptyKey)
	sensitive := c.Bool(SensitiveKey)

	var value string

	// read flag value
	if c.IsSet(ValueKey) {
		value = c.String(ValueKey)
		log.Debugf("adding flag value: (%s)", value)
	}

	// read flag file
	if value == "" && c.String(ValueFileKey) != "" {
		var err error
		if value, err = loadValueFromFile(c.String(ValueFileKey)); err != nil {
			log.Fatal("[ENVMAN] Failed to read file value: ", err)
		}
		log.Debugf("adding file flag value: (%s)", value)
	}

	// read piped stdin value
	if value == "" {
		info, err := os.Stdin.Stat()
		if err != nil {
			log.Fatalf("[ENVMAN] Failed to get standard input FileInfo: %s", err)
		}
		if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			log.Debugf("adding from piped stdin")

			configs, err := envman.GetConfigs()
			if err != nil {
				log.Fatalf("[ENVMAN] Failed to load envman config: %s", err)
			}

			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("[ENVMAN] Failed to read standard input: %s", err)
			}

			if configs.EnvBytesLimitInKB > 0 && len(data) > configs.EnvBytesLimitInKB*1024 {
				valueSizeInKB := ((float64)(len(data))) / 1024.0
				log.Warnf("environment value too large")
				log.Warnf("environment value size (%#v KB) - max allowed size: %#v KB", valueSizeInKB, (float64)(configs.EnvBytesLimitInKB))
				log.Fatalf("[ENVMAN] environment value too large - rejected")
			}

			if configs.EnvListBytesLimitInKB > 0 {
				envList, err := ReadEnvsOrCreateEmptyList(CurrentEnvStoreFilePath)
				if err != nil {
					log.Fatalf("[ENVMAN] failed to get env list, error: %s", err)
				}
				envListSizeInBytes, err := envListSizeInBytes(envList)
				if err != nil {
					log.Fatalf("[ENVMAN] failed to get env list size, error: %s", err)
				}
				if envListSizeInBytes+len(data) > configs.EnvListBytesLimitInKB*1024 {
					listSizeInKB := (float64)(envListSizeInBytes)/1024 + (float64)(len(data))/1024
					log.Warn("environment list too large")
					log.Warnf("environment list size (%#v KB) - max allowed size: %#v KB", listSizeInKB, (float64)(configs.EnvListBytesLimitInKB))
					log.Fatalf("[ENVMAN] environment list too large")
				}
			}

			value = string(data)

			log.Debugf("stdin value: (%s)", value)
		}
	}

	if err := AddEnv(CurrentEnvStoreFilePath, key, value, expand, replace, skipIfEmpty, sensitive); err != nil {
		log.Fatal("[ENVMAN] Failed to add env:", err)
	}

	log.Debugln("[ENVMAN] Env added")

	if err := logEnvs(CurrentEnvStoreFilePath); err != nil {
		log.Fatal("[ENVMAN] Failed to print:", err)
	}

	return nil
}

// AddEnv ...
func AddEnv(envStorePth string, key string, value string, expand, replace, skipIfEmpty, sensitive bool) error {
	// Load envs, or create if not exist
	environments, err := ReadEnvsOrCreateEmptyList(envStorePth)
	if err != nil {
		return err
	}

	// Validate input
	validatedValue, err := validateEnv(key, value, environments)
	if err != nil {
		return err
	}
	value = validatedValue

	// Add or update envlist
	newEnv := models.EnvironmentItemModel{
		key: value,
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			IsExpand:    pointers.NewBoolPtr(expand),
			SkipIfEmpty: pointers.NewBoolPtr(skipIfEmpty),
			IsSensitive: pointers.NewBoolPtr(sensitive),
		},
	}
	if err := newEnv.NormalizeValidateFillDefaults(); err != nil {
		return err
	}

	newEnvSlice, err := UpdateOrAddToEnvlist(environments, newEnv, replace)
	if err != nil {
		return err
	}

	return WriteEnvMapToFile(envStorePth, newEnvSlice)
}

func envListSizeInBytes(envs []models.EnvironmentItemModel) (int, error) {
	valueSizeInBytes := 0
	for _, env := range envs {
		_, value, err := env.GetKeyValuePair()
		if err != nil {
			return 0, err
		}
		valueSizeInBytes += len([]byte(value))
	}
	return valueSizeInBytes, nil
}

func validateEnv(key, value string, envList []models.EnvironmentItemModel) (string, error) {
	if key == "" {
		return "", errors.New("Key is not specified, required")
	}

	configs, err := envman.GetConfigs()
	if err != nil {
		return "", err
	}

	valueSizeInBytes := len([]byte(value))
	if configs.EnvBytesLimitInKB > 0 {
		if valueSizeInBytes > configs.EnvBytesLimitInKB*1024 {
			valueSizeInKB := ((float64)(valueSizeInBytes)) / 1024.0
			return fmt.Sprintf("environment var (%s) value is too large (%#v KB), max allowed size: %#v KB", key, valueSizeInKB, (float64)(configs.EnvBytesLimitInKB)), nil
		}
	}

	if configs.EnvListBytesLimitInKB > 0 {
		envListSizeInBytes, err := envListSizeInBytes(envList)
		if err != nil {
			return "", err
		}
		if envListSizeInBytes+valueSizeInBytes > configs.EnvListBytesLimitInKB*1024 {
			listSizeInKB := (float64)(envListSizeInBytes)/1024 + (float64)(valueSizeInBytes)/1024
			return "", fmt.Errorf("environment list is too large (%#v KB), max allowed size: %#v KB", listSizeInKB, (float64)(configs.EnvListBytesLimitInKB))
		}
	}
	return value, nil
}

func loadValueFromFile(pth string) (string, error) {
	buf, err := ioutil.ReadFile(pth)
	if err != nil {
		return "", err
	}

	str := string(buf)
	return str, nil
}

func logEnvs(envStorePth string) error {
	environments, err := ReadEnvs(envStorePth)
	if err != nil {
		return err
	}

	if len(environments) == 0 {
		log.Info("[ENVMAN] Empty envstore")
	} else {
		for _, env := range environments {
			key, value, err := env.GetKeyValuePair()
			if err != nil {
				return err
			}

			opts, err := env.GetOptions()
			if err != nil {
				return err
			}

			envString := "- " + key + ": " + value
			log.Debugln(envString)
			if !*opts.IsExpand {
				expandString := "  " + "isExpand" + ": " + "false"
				log.Debugln(expandString)
			}
		}
	}

	return nil
}
