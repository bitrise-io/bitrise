package cli

import (
	"errors"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
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
	newEnv := envman.EnvModel{Key: key, Value: value, IsExpand: expand}
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
	log.Info("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

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

	log.Info("[ENVMAN] - Env added")

	if err := printEnvs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to print:", err)
	}
}
