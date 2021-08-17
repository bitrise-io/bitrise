package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
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
		log.Fatalf("Failed to get bitrise config path, error: %s", err)
	}
	if bitriseConfigPath == "" {
		log.Fatal("No bitrise config path defined!")
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, error: %s", err)
	}

	// Normalize
	if err := bitrise.RemoveConfigRedundantFieldsAndFillStepOutputs(&bitriseConfig); err != nil {
		log.Fatalf("Failed to remove redundant fields, error: %s", err)
	}
	if err := bitrise.SaveConfigToFile(bitriseConfigPath, bitriseConfig); err != nil {
		log.Fatalf("Failed to save config to file, error: %s", err)
	}

	log.Info("Redundant fields removed")

	return nil
}
