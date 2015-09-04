package cli

import (
	"encoding/json"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/codegangsta/cli"
)

func export(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	outfilePth := c.String(OuputPathKey)
	if outfilePth == "" {
		log.Fatalln("No output file path specified!")
	}
	outFormat := c.String(OuputFormatKey)
	if outFormat == "" {
		log.Fatalln("No output file format specified!")
	}

	bitriseConfig := models.BitriseDataModel{}

	bitriseConfigBase64Data := c.String(ConfigBase64Key)
	if bitriseConfigBase64Data != "" {
		config, err := GetBitriseConfigFromBase64Data(bitriseConfigBase64Data)
		if err != nil {
			log.Fatalf("Failed to get config (bitrise.yml) from base 64 data, err: %s", err)
		}
		bitriseConfig = config
	} else {
		bitriseConfigPath, err := GetBitriseConfigFilePath(c)
		if err != nil {
			log.Fatalf("Failed to get config (bitrise.yml) path: %s", err)
		}
		if bitriseConfigPath == "" {
			log.Fatalln("Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
		}

		config, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
		if err != nil {
			log.Fatalln("Failed to validate config: ", err)
		}
		bitriseConfig = config
	}

	// serialize
	var err error
	configBytes := []byte{}
	if outFormat == "json" {
		if c.Bool(PrettyFormatKey) {
			configBytes, err = json.MarshalIndent(bitriseConfig, "", "\t")
		} else {
			configBytes, err = json.Marshal(bitriseConfig)
		}
		if err != nil {
			log.Fatalln("Failed to generate JSON: ", err)
		}
	} else if outFormat == "yaml" {
		configBytes, err = yaml.Marshal(bitriseConfig)
		if err != nil {
			log.Fatalln("Failed to generate YAML: ", err)
		}
	} else {
		log.Fatalln("Invalid output format: ", outFormat)
	}

	// write to file
	if err := fileutil.WriteBytesToFile(outfilePth, configBytes); err != nil {
		log.Fatalf("Failed to write to file (%s), error: ", err)
	}

	log.Infoln("Done, saved to path: ", outfilePth)
}
