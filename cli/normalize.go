package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/codegangsta/cli"
)

func normalize(c *cli.Context) {
	// Input validation
	bitriseConfig := models.BitriseDataModel{}
	bitriseConfigPath := ""

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	if bitriseConfigBase64Data != "" {
		config, err := GetBitriseConfigFromBase64Data(bitriseConfigBase64Data)
		if err != nil {
			log.Fatalf("Failed to get config (bitrise.yml) from base 64 data, err: %s", err)
		}
		bitriseConfig = config

		bitriseConfigPath = c.String(ConfigKey)
	} else {
		configPath, err := GetBitriseConfigFilePath(c)
		if err != nil {
			log.Fatalf("Failed to get config (bitrise.yml) path: %s", err)
		}
		bitriseConfigPath = configPath

		if bitriseConfigPath == "" {
			log.Fatalln("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
		}

		config, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
		if err != nil {
			log.Fatalln("Failed to validate config: ", err)
		}
		bitriseConfig = config
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI model version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI model's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	if bitriseConfigPath == "" {
		log.Fatal("No bitrise config path defined!")
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
