package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func normalize(c *cli.Context) {
	// Input validation
	bitriseConfigPath := c.String(ConfigKey)
	if bitriseConfigPath == "" {
		log.Fatal("No bitrise config path defined!")
	}

	// Config validation
	bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
	if err != nil {
		log.Fatalf("Failed to create bitrise cofing, err: %s", err)
	}

	// Normalize
	if err := bitrise.RemoveConfigRedundantFieldsAndFillStepOutputs(&bitriseConfig); err != nil {
		log.Fatal("Failed to remove redundant fields:", err)
	}
	if err := bitrise.SaveConfigToFile(bitriseConfigPath, bitriseConfig); err != nil {
		log.Fatal("Failed to save config to file:", err)
	}

	log.Info("Redundant fields removed")
}
