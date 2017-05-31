package cli

import (
	"encoding/json"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/urfave/cli"
)

func export(c *cli.Context) error {
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
		exportCommandHelp(c)
		log.Fatal("No output file path specified!")
	}

	if outFormat == "" {
		exportCommandHelp(c)
		log.Fatal("No output file format specified!")
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		exportCommandHelp(c)
		log.Fatalf("Failed to create bitrise config, error: %s", err)
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
			log.Fatalf("Failed to generate config JSON, error: %s", err)
		}
	} else if outFormat == output.FormatYML {
		configBytes, err = yaml.Marshal(bitriseConfig)
		if err != nil {
			log.Fatalf("Failed to generate config YML, error: %s", err)
		}
	} else {
		log.Fatalf("Invalid output format: %s", outFormat)
	}

	// write to file
	if err := fileutil.WriteBytesToFile(outfilePth, configBytes); err != nil {
		log.Fatalf("Failed to write file (%s), error: %s", outfilePth, err)
	}

	log.Infof("Done, saved to path: %s", outfilePth)

	return nil
}

func exportCommandHelp(c *cli.Context) {
	if err := cli.ShowCommandHelp(c, "export"); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}
