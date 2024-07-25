package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/bitrise/configmerge"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/urfave/cli"
)

var mergeConfigCommand = cli.Command{
	Name:      "merge",
	Usage:     "Resolves includes in a modular bitrise.yml and merges included config modules into a single bitrise.yml file.",
	ArgsUsage: "args[0]: By default, the command looks for a bitrise.yml in the current directory, custom path can be specified as an argument.",
	Action:    mergeConfig,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Output directory for the merged config file (bitrise.yml) and related config file tree (config_tree.json).",
		},
	},
}

func mergeConfig(c *cli.Context) error {
	var configPth string
	if c.Args().Present() {
		configPth = c.Args().First()
	} else {
		configPth = "bitrise.yml"
	}
	outputDir := c.String("output")

	opts := log.GetGlobalLoggerOpts()
	logger := log.NewLogger(opts)

	repoCache := configmerge.NewRepoCache()
	configReader, err := configmerge.NewConfigReader(repoCache, logger)
	if err != nil {
		return fmt.Errorf("failed to create config module reader: %w", err)
	}
	merger := configmerge.NewMerger(configReader, logger)
	mergedConfigContent, configFileTree, err := merger.MergeConfig(configPth)
	if err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	if outputDir == "" {
		if err := printOutputFiles(mergedConfigContent, *configFileTree, logger); err != nil {
			return fmt.Errorf("failed to print output files: %w", err)
		}
	} else {
		if err := writeOutputFiles(mergedConfigContent, *configFileTree, outputDir); err != nil {
			return fmt.Errorf("failed to write output files: %w", err)
		}
	}

	return nil
}

func printOutputFiles(mergedConfigContent string, configFileTree models.ConfigFileTreeModel, logger log.Logger) error {
	logger.Printf("config tree:")
	configTreeBytes, err := json.MarshalIndent(configFileTree, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse config tree: %s", err)
	}
	logger.Printf(string(configTreeBytes))

	logger.Print()
	logger.Printf("merged config:")
	logger.Printf(mergedConfigContent)

	return nil
}

func writeOutputFiles(mergedConfigContent string, configFileTree models.ConfigFileTreeModel, outputDir string) error {
	if err := pathutil.EnsureDirExist(outputDir); err != nil {
		return err
	}

	configTreeBytes, err := json.MarshalIndent(configFileTree, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to parse config tree: %s", err)
	}

	configTreePth := filepath.Join(outputDir, "config_tree.json")
	if err := fileutil.WriteBytesToFile(configTreePth, configTreeBytes); err != nil {
		return fmt.Errorf("failed to write config tree to file: %s", err)
	}

	mergedConfigPth := filepath.Join(outputDir, "bitrise.yml")
	if err := fileutil.WriteStringToFile(mergedConfigPth, mergedConfigContent); err != nil {
		return fmt.Errorf("failed to write merged config to file: %s", err)
	}

	return nil
}
