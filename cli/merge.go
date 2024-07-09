package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/bitrise/configmerge"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/urfave/cli"
)

var mergeConfigCommand = cli.Command{
	Name:      "merge",
	Usage:     "Resolves includes in a modular bitrise.yml and merge included config modules into a single bitrise.yml file.",
	ArgsUsage: "args[0]: By default the command looks for a bitrise.yml in the current directory, custom path can be specified as an argument.",
	Action:    mergeConfig,
	Flags:     []cli.Flag{},
}

func mergeConfig(c *cli.Context) error {
	var configPth string
	if c.Args().Present() {
		configPth = c.Args().First()
	} else {
		configPth = "bitrise.yml"
	}

	logger := logV2.NewLogger()
	repoInfoProvider := configmerge.NewRepoInfoProvider()
	fileReader := configmerge.NewFileReader(logger)
	fileCache := configmerge.NewFileCache(configs.GetBitriseConfigCacheDirPath(), logger)
	merger := configmerge.NewMerger(repoInfoProvider, fileReader, fileCache, logger)
	mergedConfigContent, configFileTree, err := merger.MergeConfig(configPth)
	if err != nil {
		return fmt.Errorf("failed to merge config: %s", err)
	}

	logger.Printf("config tree:")
	printConfigFileTree(*configFileTree, logger)

	logger.Println()
	logger.Printf("merge config:")
	logger.Printf(mergedConfigContent)

	return nil
}

func printConfigFileTree(configFileTree models.ConfigFileTreeModel, logger logV2.Logger) {
	b, err := json.MarshalIndent(configFileTree, "", "\t")
	if err == nil {
		logger.Printf(string(b))
	}
}
