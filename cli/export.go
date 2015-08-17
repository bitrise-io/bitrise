package cli

import (
	"encoding/json"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/codegangsta/cli"
)

func export(c *cli.Context) {
	PrintBitriseHeaderASCIIArt(c.App.Version)

	bitriseConfigPath, err := GetBitriseConfigFilePath(c)
	if err != nil {
		log.Fatalf("[BITRISE_CLI] - Failed to get config (bitrise.yml) path: %s", err)
	}
	if bitriseConfigPath == "" {
		log.Fatalln("[BITRISE_CLI] - Failed to get config (bitrise.yml) path: empty bitriseConfigPath")
	}

	outfilePth := c.String(OuputPathKey)
	if outfilePth == "" {
		log.Fatalln("No output file path specified!")
	}
	outFormat := c.String(OuputFormatKey)
	if outFormat == "" {
		log.Fatalln("No output file format specified!")
	}

	// Run workflow
	bitriseConfig, err := bitrise.ReadBitriseConfig(bitriseConfigPath)
	if err != nil {
		log.Fatalln("Failed to read config file: ", err)
	}
	log.Debugf("bitriseConfig: %#v", bitriseConfig)

	// serialize
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
