package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/versions"
	"github.com/codegangsta/cli"
)

func validate(c *cli.Context) {
	// Config validation
	bitriseConfigPath, err := GetBitriseConfigFilePath(c)
	if err != nil {
		log.Fatalf("Failed to get config (bitrise.yml) path: %s", err)
	}
	if bitriseConfigPath == "" {
		log.Fatalln("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
	}

	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		log.Fatalln("Failed to validate config: ", err)
	}

	isConfigVersionOK, err := versions.IsVersionGreaterOrEqual(models.Version, bitriseConfig.FormatVersion)
	if err != nil {
		log.Warn("bitrise CLI model version: ", models.Version)
		log.Warn("bitrise.yml Format Version: ", bitriseConfig.FormatVersion)
		log.Fatalln("Failed to compare bitrise CLI models's version with the bitrise.yml FormatVersion: ", err)
	}
	if !isConfigVersionOK {
		log.Warnf("The bitrise.yml has a higher Format Version (%s) than the bitrise CLI model's version (%s).", bitriseConfig.FormatVersion, models.Version)
		log.Fatalln("This bitrise.yml was created with and for a newer version of bitrise CLI, please upgrade your bitrise CLI to use this bitrise.yml!")
	}

	log.Info("Valid bitrise config")
}
