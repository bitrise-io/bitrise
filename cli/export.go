package cli

import (
	"encoding/json"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/output"
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
		showSubcommandHelp(c)
		failf("No output file path specified!")
	}

	if outFormat == "" {
		showSubcommandHelp(c)
		failf("No output format format specified!")
	}

	// Config validation
	bitriseConfig, warnings, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
	for _, warning := range warnings {
		log.Warnf("warning: %s", warning)
	}
	if err != nil {
		showSubcommandHelp(c)
		failf("Failed to create bitrise config, error: %s", err)
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
			failf("Failed to generate config JSON, error: %s", err)
		}
	} else if outFormat == output.FormatYML {
		node := yaml.Node{}
		if err = node.Encode(bitriseConfig); err != nil {
			failf("Failed to generate config YML, error: %s", err)
		}
		node.Style = yaml.LiteralStyle
		configBytes, err = yaml.Marshal(node)
		if err != nil {
			failf("Failed to generate config YML, error: %s", err)
		}
	} else {
		failf("Invalid output format: %s", outFormat)
	}

	// write to file
	if err := fileutil.WriteBytesToFile(outfilePth, configBytes); err != nil {
		failf("Failed to write file (%s), error: %s", outfilePth, err)
	}

	log.Infof("Done, saved to path: %s", outfilePth)

	return nil
}
