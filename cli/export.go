package cli

import (
	"encoding/json"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/codegangsta/cli"
)

func export(c *cli.Context) {
	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		log.Warn("'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

	outfilePth := c.String(OuputPathKey)
	outFormat := c.String(OuputFormatKey)
	prettyFormat := c.Bool(PrettyFormatKey)
	//

	if outfilePth == "" {
		log.Fatalln("No output file path specified!")
	}
	if outFormat == "" {
		log.Fatalln("No output file format specified!")
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		log.Fatalf("Failed to create bitrise config, err: %s", err)
	}

	// serialize
	configBytes := []byte{}
	if outFormat == output.FormatJSON {
		if prettyFormat {
			configBytes, err = json.MarshalIndent(bitriseConfig, "", "\t")
		} else {
			configBytes, err = json.Marshal(bitriseConfig)
		}
		if err != nil {
			log.Fatalln("Failed to generate JSON: ", err)
		}
	} else if outFormat == output.FormatYML {
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
