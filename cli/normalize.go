package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/codegangsta/cli"
)

func normalize(c *cli.Context) {
	// Input validation
	bitriseConfigPath, err := GetBitriseConfigFilePath(c)
	if err != nil {
		log.Fatal("Failed to get bitrise confog path:", err)
	}
	if bitriseConfigPath == "" {
		log.Fatal("No bitrise confog path defined")
	}

	// Read & validate config
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(c.App.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI version: ", c.App.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI's version (%s).", bitriseConfig.FormatVersion, c.App.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	// Normalize
	if err := bitrise.RemoveConfigRedundantFieldsAndFillStepOutputs(bitriseConfig); err != nil {
		log.Fatal("Failed to remove redundant fields:", err)
	}
	if err := bitrise.SaveConfigToFile(bitriseConfigPath, bitriseConfig); err != nil {
		log.Fatal("Failed to save config to file:", err)
	}

	log.Info("Redundant fields removed")
}
