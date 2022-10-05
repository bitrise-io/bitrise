package cli

import (
	"github.com/bitrise-io/bitrise/bitrise"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

func normalize(c *cli.Context) error {
	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		log.Warn("'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}
	//

	// Input validation
	bitriseConfigPath, err := GetBitriseConfigFilePath(bitriseConfigPath)
	if err != nil {
		failf("Failed to get bitrise config path, error: %s", err)
	}
	if bitriseConfigPath == "" {
		failf("No bitrise config path defined!")
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		failf("Failed to create bitrise config, error: %s", err)
	}

	// Normalize
	if err := bitrise.RemoveConfigRedundantFieldsAndFillStepOutputs(&bitriseConfig); err != nil {
		failf("Failed to remove redundant fields, error: %s", err)
	}
	if err := bitrise.SaveConfigToFile(bitriseConfigPath, bitriseConfig); err != nil {
		failf("Failed to save config to file, error: %s", err)
	}

	log.Info("Redundant fields removed")

	return nil
}
