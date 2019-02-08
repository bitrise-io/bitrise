package cli

import (
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/urfave/cli"
)

func unset(c *cli.Context) error {
	key := c.String(KeyKey)
	// Load envs, or create if not exist
	environments, err := envman.ReadEnvsOrCreateEmptyList()
	if err != nil {
		return err
	}

	// Add or update envlist
	newEnv := models.EnvironmentItemModel{
		key: "",
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			Unset: pointers.NewBoolPtr(true),
		},
	}

	if err := newEnv.NormalizeValidateFillDefaults(); err != nil {
		return err
	}

	newEnvSlice, err := envman.UpdateOrAddToEnvlist(environments, newEnv, true)
	if err != nil {
		return err
	}

	return envman.WriteEnvMapToFile(envman.CurrentEnvStoreFilePath, newEnvSlice)
}
