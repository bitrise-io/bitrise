package cli

import (
	"errors"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/codegangsta/cli"
)

func addEnv(key string, value string, expand, replace bool) error {
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
	newEnv := models.EnvironmentItemModel{
		key: value,
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			IsExpand: &expand,
		},
	}
	if _, err = envman.UpdateOrAddToEnvlist(environments, newEnv, replace); err != nil {
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

func logEnvs() error {
	environments, err := envman.LoadEnvMap()
	if err != nil {
		return err
	}

	if len(environments) == 0 {
		log.Info("[ENVMAN] - Empty envstore")
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

func parseExpand(c *cli.Context) bool {
	if c.IsSet(NoExpandKey) {
		return !c.Bool(NoExpandKey)
	}
	return true
}

func parseReplace(c *cli.Context) bool {
	if c.IsSet(AppendKey) {
		return !c.Bool(AppendKey)
	}
	return true
}

func addCmd(c *cli.Context) {
	log.Debugln("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	key := c.String(KeyKey)
	expand := parseExpand(c)
	replace := parseReplace(c)

	var value string
	if stdinValue != "" {
		value = stdinValue
	} else if c.IsSet(ValueKey) {
		value = c.String(ValueKey)
	} else if c.String(ValueFileKey) != "" {
		if v, err := loadValueFromFile(c.String(ValueFileKey)); err != nil {
			log.Fatal("[ENVMAN] - Failed to read file value: ", err)
		} else {
			value = v
		}
	}

	if err := addEnv(key, value, expand, replace); err != nil {
		log.Fatal("[ENVMAN] - Failed to add env:", err)
	}

	log.Debugln("[ENVMAN] - Env added")

	if err := logEnvs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to print:", err)
	}
}
