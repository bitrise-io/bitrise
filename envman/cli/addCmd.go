package cli

import (
	"errors"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func addEnv(key string, value string, expand bool) error {
	// Validate input
	if key == "" {
		return errors.New("Key is not specified, required.")
	}

	// Load envs, or create if not exist
	environments, err := envman.LoadEnvMapOrCreate()
	if err != nil {
		return err
	}

	// Add or update envlist
	newEnv := envman.EnvModel{key, value, expand}
	if _, err = envman.UpdateOrAddToEnvlist(environments, newEnv); err != nil {
		return err
	}

	return nil
}

func loadValueFromFile(pth string) (string, error) {
	buf, err := ioutil.ReadFile(pth)
	if err != nil {
		return "", err
	}

	str := string(buf)
	return str, nil
}

func addCmd(c *cli.Context) {
	log.Info("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	key := c.String(KEY_KEY)
	expand := envman.IsExpand(c.String(EXPAND_KEY))
	var value string

	if stdinValue != "" {
		value = stdinValue
	} else if c.IsSet(VALUE_KEY) {
		value = c.String(VALUE_KEY)
	} else if c.String(VALUE_FILE_KEY) != "" {
		if v, err := loadValueFromFile(c.String(VALUE_FILE_KEY)); err != nil {
			log.Fatal("[ENVMAN] - Failed to read file value: ", err)
		} else {
			value = v
		}
	}

	if err := addEnv(key, value, expand); err != nil {
		log.Fatal("[ENVMAN] - Failed to add env:", err)
	}

	log.Info("[ENVMAN] - Env added")

	if err := printEnvs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to print:", err)
	}
}
